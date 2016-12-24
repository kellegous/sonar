package config

import (
	"bytes"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	goods := map[string]time.Duration{
		"400":   400 * time.Nanosecond,
		"400ns": 400 * time.Nanosecond,
		"3ms":   3 * time.Millisecond,
		"200µs": 200 * time.Microsecond,
		"3s":    3 * time.Second,
		"3sec":  3 * time.Second,
		"10min": 10 * time.Minute,
		"10m":   10 * time.Minute,
		"6h":    6 * time.Hour,
		"60hr":  60 * time.Hour,
	}

	for s, exp := range goods {
		d, err := parseDuration(s)
		if err != nil {
			t.Fatalf("parsed failed %s: %s", s, err)
		}

		if d != exp {
			t.Fatalf("for %s, expected %s got %s", s, exp, d)
		}
	}

	bads := []string{
		"400hours",
		"40.3ms",
		"butter",
	}

	for _, bad := range bads {
		if _, err := parseDuration(bad); err == nil {
			t.Fatalf("expected err for %s", bad)
		}
	}
}

func TestDefaults(t *testing.T) {
	txt := `
	[[hosts]]
	ip = "127.0.0.1"
	name = "localhost"
	`

	var cfg Config
	if err := cfg.Read(bytes.NewBufferString(txt)); err != nil {
		t.Fatal(err)
	}

	if cfg.Addr != defaultAddr {
		t.Fatalf("expected addr of %s, got %s", defaultAddr, cfg.Addr)
	}

	if cfg.SamplePeriod != defaultSamplePeriod {
		t.Fatalf("expected sample-period of %s got %s",
			defaultSamplePeriod,
			cfg.SamplePeriod)
	}

	if cfg.SamplesPerPeriod != defaultSamplesPerPeriod {
		t.Fatalf("expected samples-per-period of %d got %d",
			defaultSamplesPerPeriod,
			cfg.SamplesPerPeriod)
	}
}

func TestRead(t *testing.T) {
	txt := `
	addr = ":80"
	sample-period = "45s"
	samples-per-period = 120

	[[hosts]]
	ip = "127.0.0.1"
	name = "localhost"
	[[hosts]]
	ip = "8.8.8.8"
	name = "Google"
	`

	var cfg Config
	if err := cfg.Read(bytes.NewBufferString(txt)); err != nil {
		t.Fatal(err)
	}

	if cfg.Addr != ":80" {
		t.Fatalf("expected addr of :80 got %s", cfg.Addr)
	}

	if cfg.SamplePeriod != 45*time.Second {
		t.Fatalf("expected sample-period of 45s got %s",
			cfg.SamplePeriod)
	}

	if cfg.SamplesPerPeriod != 120 {
		t.Fatalf("expected samples-per-period of 120 got %d",
			cfg.SamplesPerPeriod)
	}

	if len(cfg.Hosts) != 2 {
		t.Fatalf("expected 2 hosts got %d", len(cfg.Hosts))
	}

	if cfg.Hosts[0].IP != "127.0.0.1" || cfg.Hosts[0].Name != "localhost" {
		t.Fatalf("invalid Hosts[0] = %v", cfg.Hosts[0])
	}

	if cfg.Hosts[1].IP != "8.8.8.8" || cfg.Hosts[1].Name != "Google" {
		t.Fatalf("invalid hosts[1] = %v", cfg.Hosts[1])
	}
}
