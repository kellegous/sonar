package store

import (
	"testing"
	"time"

	"sonar/ping"
)

func TestResultMarshal(t *testing.T) {
	r := &ping.Results{
		Data: []time.Duration{
			time.Second,
			23 * time.Hour,
			300 * time.Millisecond,
		},
	}

	b := marshalResults(r)

	rp := unmarshalResults(b)

	if len(rp.Data) != len(r.Data) {
		t.Fatalf("len is wrong %d vs %d", len(rp.Data), len(r.Data))
	}

	for i, dur := range r.Data {
		if dur != rp.Data[i] {
			t.Fatalf("data is wrong at %d: %s vs %s",
				i, dur.String(), rp.Data[i].String())
		}
	}
}
