package web

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed ui
var assets embed.FS

const assetsPath = "pkg/web/ui"

func startWebPackWatch(
	ctx context.Context,
	root string,
) error {
	c := exec.Command("npx",
		"webpack",
		"watch",
		"--mode=development")
	c.Dir = root
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Start()
}

func getContent(
	ctx context.Context,
	opts *Options,
) (http.FileSystem, error) {
	if opts.UseDevRoot == "" {
		fs, err := fs.Sub(assets, "ui")
		if err != nil {
			return nil, err
		}

		return http.FS(fs), nil
	}

	if err := startWebPackWatch(ctx, opts.UseDevRoot); err != nil {
		return nil, err
	}
	return http.Dir(filepath.Join(opts.UseDevRoot, assetsPath)), nil
}
