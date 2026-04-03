package web

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/kellegous/glue/metrics"
	"github.com/kellegous/sonar/internal/config"
	"github.com/kellegous/sonar/internal/store"
	"github.com/kellegous/sonar/sonar_connect"
)

const rpcPrefix = "/rpc"

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

	path, handler := sonar_connect.NewSonarHandler(
		svr,
		connect.WithInterceptors(metrics.ForRPC()),
	)
	m.Handle(rpcPrefix+path, http.StripPrefix(rpcPrefix, handler))

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

	return http.ListenAndServe(cfg.Addr, metrics.ForHTTP(m))
}
