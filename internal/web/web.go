package web

import (
	"context"
	"net/http"

	"github.com/kellegous/sonar"
	"github.com/kellegous/sonar/internal/config"
	"github.com/kellegous/sonar/internal/store"
)

func ListenAndServe(
	ctx context.Context,
	cfg *config.Config,
	s *store.Store,
	assets http.Handler,
) error {
	m := http.NewServeMux()

	svr := &server{
		cfg:   cfg,
		store: s,
	}

	m.Handle(sonar.SonarPathPrefix, sonar.NewSonarServer(svr))

	m.HandleFunc(
		"/api/v1/current",
		func(w http.ResponseWriter, r *http.Request) {
			apiCurrent(w, r, svr)
		})

	m.HandleFunc(
		"/api/v1/hourly",
		func(w http.ResponseWriter, r *http.Request) {
			apiByHour(w, r, svr)
		})

	m.Handle("/", assets)

	return http.ListenAndServe(cfg.Addr, m)
}
