package build

import (
	"errors"
	"runtime/debug"
	"time"

	"github.com/kellegous/buildname"
)

type Summary struct {
	SHA        string    `json:"sha"`
	CommitTime time.Time `json:"commit_time"`
	Name       string    `json:"name"`
	Go         GoInfo    `json:"go"`
}

type GoInfo struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

func ReadSummary() (*Summary, error) {
	b, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("build info unavailable")
	}

	return summaryFrom(b)
}

func summaryFrom(info *debug.BuildInfo) (*Summary, error) {
	settings := map[string]string{}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}

	sha := settings["vcs.revision"]

	ct, err := time.Parse(time.RFC3339, settings["vcs.time"])
	if err != nil {
		return nil, err
	}

	return &Summary{
		SHA:        sha,
		CommitTime: ct,
		Name:       buildname.FromVersion(sha),
		Go: GoInfo{
			Module:  info.Main.Path,
			Version: info.GoVersion,
			OS:      settings["GOOS"],
			Arch:    settings["GOARCH"],
		},
	}, nil
}
