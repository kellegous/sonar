package web

import (
	"net/http"

	"sonar/config"
	"sonar/store"

	"github.com/kellegous/pork"
)

// ListenAndServe ...
func ListenAndServe(cfg *config.Config, s *store.Store) error {
	r := pork.NewRouter(nil, nil, nil)

	setupAPI(r, cfg, s)

	return http.ListenAndServe(cfg.Addr, r)
}
