package web

import (
	"context"
	"net"
	"sort"
	"sync"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kellegous/sonar"
	"github.com/kellegous/sonar/internal/config"
	"github.com/kellegous/sonar/internal/store"
	"github.com/kellegous/sonar/sonar_connect"
)

type server struct {
	cfg   *config.Config
	store *store.Store
}

var _ sonar_connect.SonarHandler = (*server)(nil)

func buildStats(t time.Time, vals []time.Duration) *sonar.Stats {
	stats := &sonar.Stats{
		Time: timestamppb.New(t),
	}

	n := len(vals)
	if n == 0 {
		return stats
	}

	// filter out any lost packets
	valid := make([]int, 0, n)
	for _, d := range vals {
		v := int(d.Nanoseconds())
		if v == 0 {
			continue
		}
		valid = append(valid, v)
	}

	// TODO(knorton): should this be len(valid)?
	stats.Count = uint32(n)
	stats.Loss = 1.0 - float64(len(valid))/float64(n)

	// if all packets were lost, we can't do anything else.
	if len(valid) == 0 {
		return stats
	}

	sort.Ints(valid)
	max := 0
	min := 0x7fffffffffffffff

	mu := 0.0
	for _, d := range valid {
		mu += float64(d)
		if d > max {
			max = d
		}
		if d < min {
			min = d
		}
	}
	mu /= float64(len(valid))

	stats.Avg = mu
	stats.Max = uint32(max)
	stats.Min = uint32(min)
	stats.P10 = uint32(perc(0.1, valid))
	stats.P90 = uint32(perc(0.9, valid))
	stats.P50 = uint32(perc(0.5, valid))
	return stats
}

func (s *server) GetAll(
	ctx context.Context,
	req *connect.Request[sonar.GetAllRequest],
) (*connect.Response[sonar.GetAllResponse], error) {
	msg := req.Msg
	g, ctx := errgroup.WithContext(ctx)

	var current *sonar.GetCurrentResponse
	g.Go(func() error {
		var err error
		res, err := s.GetCurrent(
			ctx,
			connect.NewRequest(&emptypb.Empty{}))
		if err != nil {
			return err
		}
		current = res.Msg
		return nil
	})

	var hourly *sonar.GetHourlyResponse
	g.Go(func() error {
		var err error
		res, err := s.GetHourly(
			ctx,
			connect.NewRequest(&sonar.GetHourlyRequest{Hours: msg.GetHours()}))
		if err != nil {
			return err
		}
		hourly = res.Msg
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// TODO(kellegous): This is kind of dumb and I should clean it up.
	hosts := make([]*sonar.GetAllResponse_HostStats, 0, len(s.cfg.Hosts))
	for i, host := range current.Hosts {
		hosts = append(hosts, &sonar.GetAllResponse_HostStats{
			Host:    host.Host,
			Hours:   hourly.Hosts[i].Hours,
			Current: host.Stats,
		})
	}

	return connect.NewResponse(&sonar.GetAllResponse{
		Hosts: hosts,
	}), nil
}

func (s *server) GetCurrent(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[sonar.GetCurrentResponse], error) {
	cfg := s.cfg
	hosts := make([]*sonar.GetCurrentResponse_HostStats, 0, len(cfg.Hosts))
	for _, host := range cfg.Hosts {
		t, vals, err := s.store.Current(host.IP)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}

		c := &sonar.GetCurrentResponse_HostStats{
			Host: &sonar.Host{
				Ip:   host.IP.String(),
				Name: host.Name,
			},
			Stats: buildStats(t, vals),
		}

		hosts = append(hosts, c)
	}

	return connect.NewResponse(&sonar.GetCurrentResponse{Hosts: hosts}), nil
}

func (s *server) GetHourly(
	ctx context.Context,
	req *connect.Request[sonar.GetHourlyRequest],
) (*connect.Response[sonar.GetHourlyResponse], error) {
	msg := req.Msg
	cfg := s.cfg

	nHrs := int(msg.GetHours())
	st := time.Now().Add(-time.Duration(nHrs-1) * time.Hour).Truncate(time.Hour)

	hours := make([]*sonar.GetHourlyResponse_HostStats, 0, len(cfg.Hosts))
	for _, host := range cfg.Hosts {
		data := make([][]time.Duration, nHrs)
		if err := s.store.ForEach(
			store.NewMarker(host.IP, st),
			store.NewMarker(host.IP, store.Last),
			func(ip net.IP, t time.Time, vals []time.Duration) error {
				ix := int(t.Sub(st).Nanoseconds() / int64(time.Hour))
				if ix < 0 || ix > nHrs {
					return nil
				}
				data[ix] = append(data[ix], vals...)
				return nil
			},
		); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}

		curr := &sonar.GetHourlyResponse_HostStats{
			Host: &sonar.Host{
				Ip:   host.IP.String(),
				Name: host.Name,
			},
			Hours: make([]*sonar.Stats, nHrs),
		}

		for ix, vals := range data {
			curr.Hours[ix] = buildStats(st.Add(time.Duration(ix)*time.Hour), vals)
		}

		hours = append(hours, curr)
	}

	return connect.NewResponse(&sonar.GetHourlyResponse{Hosts: hours}), nil
}

func (s *server) GetStoreStats(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[sonar.GetStoreStatsResponse], error) {
	var lock sync.Mutex
	earliest := store.Last
	latest := store.First
	var count int64
	g, _ := errgroup.WithContext(ctx)
	for _, host := range s.cfg.Hosts {
		host := host
		g.Go(func() error {
			var a, b time.Time
			var c int64
			if err := s.store.ForEach(
				store.NewMarker(host.IP, store.First),
				store.NewMarker(host.IP, store.Last),
				func(ip net.IP, t time.Time, vals []time.Duration) error {
					if a.IsZero() || t.Before(a) {
						a = t
					}
					if b.IsZero() || t.After(b) {
						b = t
					}
					c += int64(len(vals))
					return nil
				}); err != nil {
				return err
			}

			lock.Lock()
			defer lock.Unlock()
			if a.Before(earliest) {
				earliest = a
			}
			if b.After(latest) {
				latest = b
			}
			count += c
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&sonar.GetStoreStatsResponse{
		Count:    count,
		Earliest: timestamppb.New(earliest),
		Latest:   timestamppb.New(latest),
	}), nil
}
