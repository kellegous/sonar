package config

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/kellegous/glue/fn"
)

const (
	DefaultSamplePeriod     = 30 * time.Second
	DefaultAddr             = ":4065"
	DefaultSamplesPerPeriod = 10
	DefaultDataPath         = "data"
)

type Config struct {
	DataPath         string        `toml:"data-path"`
	SamplesPerPeriod int           `toml:"samples-per-period"`
	SamplePeriod     time.Duration `toml:"sample-period"`
	Hosts            []*Host       `toml:"hosts"`
	Addr             string        `toml:"addr"`
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
}

func (c *Config) ReadFile(filename string) (err error) {
	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fn.WithCare(r.Close, &err)

	return c.Read(r, filepath.Dir(filename))
}

func (c *Config) Read(r io.Reader, base string) error {
	if _, err := toml.NewDecoder(r).Decode(c); err != nil {
		return err
	}

	c.applyDefaults(base)

	return nil
}
