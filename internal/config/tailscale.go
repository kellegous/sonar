package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultTailscaleHostname = "sonar"
	DefaultTailscaleStateDir = "tailscale"
	TailscaleAuthKeyEnvKey   = "TAILSCALE_AUTHKEY"
)

type Tailscale struct {
	AuthKey  string `toml:"auth-key"`
	Hostname string `toml:"hostname"`
	StateDir string `toml:"state-dir"`
}

func (t *Tailscale) applyDefaults(base string) {
	if t.AuthKey == "" {
		t.AuthKey = os.Getenv(TailscaleAuthKeyEnvKey)
	}

	if t.Hostname == "" {
		t.Hostname = DefaultTailscaleHostname
	}

	if t.StateDir == "" {
		t.StateDir = DefaultTailscaleStateDir
	}

	if !filepath.IsAbs(t.StateDir) {
		t.StateDir = filepath.Join(base, t.StateDir)
	}
}
