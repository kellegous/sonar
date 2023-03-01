package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kellegous/sonar/pkg/config"
	"github.com/kellegous/sonar/pkg/logging"
	"github.com/kellegous/sonar/pkg/store"
	"github.com/kellegous/tsweb"
	"tailscale.com/tsnet"
)

type Server struct {
	Config *config.Config
	Store  *store.Store
}

func (s *Server) ListenAndServe(
	ctx context.Context,
	opts *Options,
) error {
	m := http.NewServeMux()

	content, err := getContent(ctx, opts)
	if err != nil {
		return err
	}

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
	m.Handle("/", http.FileServer(content))

	if s.Config.UseTailscale() {
		return listenAndServeViaTailscale(ctx, &s.Config.Tailscale, m)
	}
	return http.ListenAndServe(s.Config.Addr, m)
}

func listenAndServeViaTailscale(
	ctx context.Context,
	cfg *config.Tailscale,
	h http.Handler,
) error {
	svc, err := tsweb.Start(&tsnet.Server{
		AuthKey:  cfg.AuthKey,
		Hostname: cfg.Hostname,
		Dir:      cfg.StateDir,
		Logf: func(format string, args ...any) {
			logging.L(ctx).Info(fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		return err
	}
	defer svc.Close()

	l, err := svc.ListenTLS("tcp", ":https")
	if err != nil {
		return err
	}
	defer l.Close()

	ch := make(chan error, 1)
	go func() {
		ch <- svc.RedirectHTTP(ctx)
	}()
	go func() {
		ch <- http.Serve(l, h)
	}()

	return <-ch
}
