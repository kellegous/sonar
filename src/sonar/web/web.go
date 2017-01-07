package web

import (
	"bytes"
	"net/http"

	"sonar/config"
	"sonar/store"
	"sonar/web/internal"

	"github.com/kellegous/pork"
)

type content struct{}

func (c *content) ServePork(w pork.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len(w.ServedFromPrefix()):]

	if path == "" || path[len(path)-1] == '/' {
		path += "index.html"
	}

	n, err := internal.AssetInfo(path)
	if err != nil {
		w.ServeNotFound()
	}

	a, err := internal.Asset(path)
	if err != nil {
		w.ServeNotFound()
	}

	http.ServeContent(w, r, n.Name(), n.ModTime(), bytes.NewReader(a))
}

func contentFrom(assetsDir string) pork.Responder {
	if assetsDir != "" {
		return pork.Content(
			pork.NewConfig(pork.None),
			http.Dir(assetsDir))
	}

	return &content{}
}

// ListenAndServe ...
func ListenAndServe(cfg *config.Config, s *store.Store, assetsDir string) error {
	r := pork.NewRouter(nil, nil, nil)

	setupAPI(r, cfg, s)

	r.RespondWith("/", contentFrom(assetsDir))

	return http.ListenAndServe(cfg.Addr, r)
}
