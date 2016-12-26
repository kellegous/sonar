package web

import (
	"net/http"

	"sonar/config"
	"sonar/store"

	"github.com/kellegous/pork"
)

func contentFrom(assetsDir string) pork.Responder {
	return pork.Content(pork.NewConfig(pork.None),
		http.Dir(assetsDir))
}

// ListenAndServe ...
func ListenAndServe(cfg *config.Config, s *store.Store, assetsDir string) error {
	r := pork.NewRouter(nil, nil, nil)

	setupAPI(r, cfg, s)

	r.RespondWith("/", contentFrom(assetsDir))

	return http.ListenAndServe(cfg.Addr, r)
}
