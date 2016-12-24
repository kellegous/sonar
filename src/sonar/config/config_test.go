package config

import (
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
