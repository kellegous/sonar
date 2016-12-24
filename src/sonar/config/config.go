package config

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
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
	Addr             string `toml:"addr"`
	SamplesPerPeriod int    `toml:"samples-per-period"`
	SamplePeriodStr  string `toml:"sample-period"`
	Hosts            []*Host
	sampleFreq       time.Duration
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

func expand(c *Config) error {
	d, err := parseDuration(c.SamplePeriodStr)
	if err != nil {
		return err
	}
	c.sampleFreq = d

	return nil
}

// ReadFile ...
func (c *Config) ReadFile(filename string) error {
	if _, err := toml.DecodeFile(filename, c); err != nil {
		return err
	}
	return nil
}

// Read ...
func (c *Config) Read(r io.Reader) error {
	if _, err := toml.DecodeReader(r, c); err != nil {
		return err
	}
	return nil
}
