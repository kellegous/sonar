package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets
var assets embed.FS

func Assets() (http.Handler, error) {
	d, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil, err
	}

	return http.FileServer(http.FS(d)), nil
}
