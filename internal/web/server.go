package web

import (
	"context"
	"net"
	"net/http"
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

type Server struct {
	Config *config.Config
	Store  *store.Store
}

var _ sonar.Sonar = (*Server)(nil)

func (s *Server) ListenAndServe(
	ctx context.Context,
	assets http.Handler,
) error {
	m := http.NewServeMux()

	m.Handle(sonar.SonarPathPrefix, sonar.NewSonarServer(s))

	m.HandleFunc(
		"/api/v1/current",
		func(w http.ResponseWriter, r *http.Request) {
			apiCurrent(w, r, s)
		})

	m.HandleFunc(
		"/api/v1/hourly",
		func(w http.ResponseWriter, r *http.Request) {
			apiByHour(w, r, s)
		})

	m.Handle("/", assets)

	return http.ListenAndServe(s.Config.Addr, m)
}

func (s *Server) GetCurrent(ctx context.Context, req *emptypb.Empty) (*sonar.GetCurrentResponse, error) {
	return nil, twirp.NewError(twirp.Unimplemented, "not implemented")
}

func (s *Server) GetHourly(ctx context.Context, req *emptypb.Empty) (*sonar.GetHourlyResponse, error) {
	return nil, twirp.NewError(twirp.Unimplemented, "not implemented")
}

func (s *Server) GetStoreStats(ctx context.Context, req *emptypb.Empty) (*sonar.GetStoreStatsResponse, error) {
	var lock sync.Mutex
	earliest := store.Last
	latest := store.First
	var count int64
	g, _ := errgroup.WithContext(ctx)
	for _, host := range s.Config.Hosts {
		host := host
		g.Go(func() error {
			var a, b time.Time
			var c int64
			if err := s.Store.ForEach(
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
