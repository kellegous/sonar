package store

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResultMarshal(t *testing.T) {
	r := []time.Duration{
		time.Second,
		23 * time.Hour,
		300 * time.Millisecond,
	}

	b := marshalResults(r)

	rp := unmarshalResults(b)

	if len(rp) != len(r) {
		t.Fatalf("len is wrong %d vs %d", len(rp), len(r))
	}

	for i, dur := range r {
		if dur != rp[i] {
			t.Fatalf("data is wrong at %d: %s vs %s",
				i, dur.String(), rp[i].String())
		}
	}
}

func newStore(t *testing.T) (*Store, func() error) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	s, err := Open(filepath.Join(tmp, "db"))
	if err != nil {
		os.RemoveAll(tmp)
		t.Fatal(err)
	}

	return s, func() error {
		return os.RemoveAll(tmp)
	}
}

func intArraysAreSame(t *testing.T, a, b []int) {
	if len(a) != len(b) {
		t.Fatalf("len(%v) != len(%v)", a, b)
	}

	for i, n := 0, len(a); i < n; i++ {
		if a[i] != b[i] {
			t.Fatalf("not equal: %v vs %v", a, b)
		}
	}
}

func readDoesProduce(t *testing.T, s *Store, fr, to *Marker, exp []int) {
	var got []int
	if err := s.ForEach(fr, to, func(ip net.IP, tm time.Time, r []time.Duration) error {
		ix := int(tm.UnixNano())
		if ix != int(r[0]) {
			t.Fatalf("Invalid results %d instead of %d", int(r[0]), ix)
		}
		got = append(got, ix)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	intArraysAreSame(t, exp, got)
}

func TestReadWrite(t *testing.T) {
	ipA := net.IP([]byte{8, 8, 8, 8})
	ipB := net.IP([]byte{4, 4, 2, 2})
	ipC := net.IP([]byte{0xff, 0xff, 0xff, 0xff})

	ips := []net.IP{ipA, ipB, ipC}

	s, cleanup := newStore(t)
	defer cleanup()

	for _, ip := range ips {
		for i := 1; i <= 10; i++ {
			d := time.Duration(uint64(i))
			if err := s.Write(ip, time.Unix(0, int64(i)), []time.Duration{d}); err != nil {
				t.Fatal(err)
			}
		}
	}

	for _, ip := range ips {
		readDoesProduce(t, s,
			NewMarker(ip, First),
			NewMarker(ip, Last),
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
		readDoesProduce(t, s,
			NewMarker(ip, Last),
			NewMarker(ip, First),
			[]int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1})
		readDoesProduce(t, s,
			NewMarker(ip, time.Unix(0, 1)),
			NewMarker(ip, time.Unix(0, 1)),
			[]int{1})
		readDoesProduce(t, s,
			NewMarker(ip, time.Unix(0, 1)),
			NewMarker(ip, time.Unix(0, 3)),
			[]int{1, 2, 3})
		readDoesProduce(t, s,
			NewMarker(ip, time.Unix(0, 3)),
			NewMarker(ip, time.Unix(0, 1)),
			[]int{3, 2, 1})
		readDoesProduce(t, s,
			NewMarker(ip, time.Unix(0, 100)),
			NewMarker(ip, time.Unix(0, 200)),
			nil)
		readDoesProduce(t, s,
			NewMarker(ip, time.Unix(0, 200)),
			NewMarker(ip, time.Unix(0, 100)),
			nil)
		readDoesProduce(t, s,
			NewMarker(ip, First),
			NewMarker(ip, First),
			nil)
		readDoesProduce(t, s,
			NewMarker(ip, Last),
			NewMarker(ip, Last),
			nil)
	}
}
