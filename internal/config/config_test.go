package config

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	txt := `
	[[hosts]]
	ip = "127.0.0.1"
	name = "localhost"
	`

	var cfg Config
	if err := cfg.Read(bytes.NewBufferString(txt), "/"); err != nil {
		t.Fatal(err)
	}

	if cfg.Addr != DefaultAddr {
		t.Fatalf("expected addr of %s, got %s", DefaultAddr, cfg.Addr)
	}

	if cfg.SamplePeriod != DefaultSamplePeriod {
		t.Fatalf("expected sample-period of %s got %s",
			DefaultSamplePeriod,
			cfg.SamplePeriod)
	}

	if cfg.SamplesPerPeriod != DefaultSamplesPerPeriod {
		t.Fatalf("expected samples-per-period of %d got %d",
			DefaultSamplesPerPeriod,
			cfg.SamplesPerPeriod)
	}
}

func ipEquals(ip net.IP, v string) bool {
	return ip.Equal(net.ParseIP(v))
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
	if err := cfg.Read(bytes.NewBufferString(txt), "/"); err != nil {
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

	if !ipEquals(cfg.Hosts[0].IP, "127.0.0.1") || cfg.Hosts[0].Name != "localhost" {
		t.Fatalf("invalid Hosts[0] = %v", cfg.Hosts[0])
	}

	if !ipEquals(cfg.Hosts[1].IP, "8.8.8.8") || cfg.Hosts[1].Name != "Google" {
		t.Fatalf("invalid hosts[1] = %v", cfg.Hosts[1])
	}
}
