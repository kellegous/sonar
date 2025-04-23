package web

import (
	"context"
	"net/http"

	"github.com/kellegous/sonar"
	"github.com/kellegous/sonar/internal/config"
	"github.com/kellegous/sonar/internal/store"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/types/known/emptypb"
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
	return nil, twirp.NewError(twirp.Unimplemented, "not implemented")
}
