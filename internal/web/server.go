package web

import (
	"context"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/twitchtv/twirp"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/kellegous/sonar"
	"github.com/kellegous/sonar/internal/config"
	"github.com/kellegous/sonar/internal/store"
)

type server struct {
	cfg   *config.Config
	store *store.Store
}

var _ sonar.Sonar = (*server)(nil)

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

func (s *server) GetCurrent(
	ctx context.Context,
	req *emptypb.Empty,
) (*sonar.GetCurrentResponse, error) {
	cfg := s.cfg
	hosts := make([]*sonar.GetCurrentResponse_HostStats, 0, len(cfg.Hosts))
	for _, host := range cfg.Hosts {
		t, vals, err := s.store.Current(host.IP)
		if err != nil {
			return nil, twirp.InternalErrorWith(err)
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

	return &sonar.GetCurrentResponse{Hosts: hosts}, nil
}

func (s *server) GetHourly(
	ctx context.Context,
	req *sonar.GetHourlyRequest,
) (*sonar.GetHourlyResponse, error) {
	return nil, twirp.NewError(twirp.Unimplemented, "not implemented")
}

func (s *server) GetStoreStats(
	ctx context.Context,
	req *emptypb.Empty,
) (*sonar.GetStoreStatsResponse, error) {
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
		return nil, twirp.InternalErrorWith(err)
	}

	return &sonar.GetStoreStatsResponse{
		Count:    count,
		Earliest: timestamppb.New(earliest),
		Latest:   timestamppb.New(latest),
	}, nil
}
