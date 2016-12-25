package config

import (
	"io"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	defaultSamplePeriod     = 20 * time.Second
	defaultAddr             = ":7699"
	defaultSamplesPerPeriod = 10
	defaultDataPath         = "data"
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
	IP   net.IP
	Name string
}

// Config ...
type Config struct {
	DataPath         string
	Addr             string
	SamplesPerPeriod int
	SamplePeriod     time.Duration
	Hosts            []*Host
}

type decl struct {
	DataPath         string `toml:"data-path"`
	Addr             string `toml:"addr"`
	SamplesPerPeriod int    `toml:"samples-per-period"`
	SamplePeriod     string `toml:"sample-period"`
	Hosts            []*struct {
		IP   string `toml:"ip"`
		Name string `toml:"name"`
	} `toml:"hosts"`
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

func (c *Config) apply(root string, d *decl) error {
	c.Addr = d.Addr
	c.SamplesPerPeriod = d.SamplesPerPeriod

	hosts := make([]*Host, 0, len(d.Hosts))
	for _, host := range d.Hosts {
		hosts = append(hosts, &Host{
			IP:   net.ParseIP(host.IP),
			Name: host.Name,
		})
	}
	c.Hosts = hosts

	if d.DataPath == "" {
		d.DataPath = defaultDataPath
	}

	if !filepath.IsAbs(d.DataPath) {
		p, err := filepath.Abs(filepath.Join(root, d.DataPath))
		if err != nil {
			return err
		}
		c.DataPath = p
	} else {
		c.DataPath = d.DataPath
	}

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
	return c.apply(filepath.Dir(filename), &d)
}

// Read ...
func (c *Config) Read(r io.Reader) error {
	var d decl
	if _, err := toml.DecodeReader(r, &d); err != nil {
		return err
	}
	return c.apply("/", &d)
}
