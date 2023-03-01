package config

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	DefaultSamplePeriod     = 30 * time.Second
	DefaultAddr             = ":7699"
	DefaultSamplesPerPeriod = 10
	DefaultDataPath         = "data"
)

type Config struct {
	DataPath         string        `toml:"data-path"`
	SamplesPerPeriod int           `toml:"samples-per-period"`
	SamplePeriod     time.Duration `toml:"sample-period"`
	Hosts            []*Host       `toml:"hosts"`
	Addr             string        `toml:"addr"`
	Tailscale        Tailscale     `toml:"tailscale"`
}

func (c *Config) UseTailscale() bool {
	return c.Tailscale.AuthKey != ""
}

func (c *Config) applyDefaults(base string) {
	if c.DataPath == "" {
		c.DataPath = DefaultDataPath
	}

	if c.Addr == "" {
		c.Addr = DefaultAddr
	}

	if c.SamplesPerPeriod == 0 {
		c.SamplesPerPeriod = DefaultSamplesPerPeriod
	}

	if c.SamplePeriod == 0 {
		c.SamplePeriod = DefaultSamplePeriod
	}

	if !filepath.IsAbs(c.DataPath) {
		c.DataPath = filepath.Join(base, c.DataPath)
	}

	c.Tailscale.applyDefaults(base)
}

func (c *Config) ReadFile(filename string) error {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer r.Close()

	return c.Read(r, filepath.Dir(filename))
}

func (c *Config) Read(r io.Reader, base string) error {
	if _, err := toml.NewDecoder(r).Decode(c); err != nil {
		return err
	}

	c.applyDefaults(base)

	return nil
}
