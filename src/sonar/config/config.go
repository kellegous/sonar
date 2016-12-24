package config

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	defaultSamplePeriod     = 20 * time.Second
	defaultAddr             = ":7699"
	defaultSamplesPerPeriod = 10
)

type unit struct {
	p string
	d time.Duration
}

var units = []*unit{
	&unit{"sec", time.Second},
	&unit{"min", time.Minute},
	&unit{"ms", time.Millisecond},
	&unit{"µs", time.Microsecond},
	&unit{"ns", time.Nanosecond},
	&unit{"hr", time.Hour},
	&unit{"s", time.Second},
	&unit{"m", time.Minute},
	&unit{"h", time.Hour},
}

// Host ...
type Host struct {
	IP   string `toml:"ip"`
	Name string `toml:"name"`
}

// Config ...
type Config struct {
	Addr             string
	SamplesPerPeriod int
	SamplePeriod     time.Duration
	Hosts            []*Host
}

type decl struct {
	Addr             string  `toml:"addr"`
	SamplesPerPeriod int     `toml:"samples-per-period"`
	SamplePeriod     string  `toml:"sample-period"`
	Hosts            []*Host `toml:"hosts"`
}

func parseDuration(s string) (time.Duration, error) {
	m := time.Nanosecond
	for _, unit := range units {
		if strings.HasSuffix(s, unit.p) {
			m = unit.d
			s = s[:len(s)-len(unit.p)]
			break
		}
	}

	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Duration(0), err
	}

	return time.Duration(v) * m, nil
}

func (c *Config) apply(d *decl) error {
	c.Addr = d.Addr
	c.SamplesPerPeriod = d.SamplesPerPeriod
	c.Hosts = d.Hosts

	if d.SamplePeriod != "" {
		sp, err := parseDuration(d.SamplePeriod)
		if err != nil {
			return err
		}
		c.SamplePeriod = sp
	} else {
		c.SamplePeriod = time.Duration(0)
	}

	if c.Addr == "" {
		c.Addr = defaultAddr
	}

	if c.SamplesPerPeriod == 0 {
		c.SamplesPerPeriod = defaultSamplesPerPeriod
	}

	if c.SamplePeriod == 0 {
		c.SamplePeriod = defaultSamplePeriod
	}

	return nil
}

// ReadFile ...
func (c *Config) ReadFile(filename string) error {
	var d decl
	if _, err := toml.DecodeFile(filename, &d); err != nil {
		return err
	}
	return c.apply(&d)
}

// Read ...
func (c *Config) Read(r io.Reader) error {
	var d decl
	if _, err := toml.DecodeReader(r, &d); err != nil {
		return err
	}
	return c.apply(&d)
}
