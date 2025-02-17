package web

import (
	"context"
	"net/http"

	"github.com/kellegous/sonar/pkg/config"
	"github.com/kellegous/sonar/pkg/store"
)

type Server struct {
	Config *config.Config
	Store  *store.Store
}

func (s *Server) ListenAndServe(
	ctx context.Context,
	assets http.Handler,
) error {
	m := http.NewServeMux()

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
