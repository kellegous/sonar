package store

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	// First ...
	First = time.Unix(0, 0)

	// Last ...
	Last = time.Unix(0, 0x7fffffffffffffff)

	// ErrStop ...
	ErrStop = errors.New("stop")
)

// Store ...
type Store struct {
	db *leveldb.DB
}

// Marker ...
type Marker struct {
	b []byte
}

func isFirst(m *Marker) bool {
	return m.b[8] == 0 &&
		m.b[9] == 0 &&
		m.b[10] == 0 &&
		m.b[11] == 0 &&
		m.b[12] == 0 &&
		m.b[13] == 0 &&
		m.b[14] == 0 &&
		m.b[15] == 0
}

func isLast(m *Marker) bool {
	return m.b[8] == 0x7f &&
		m.b[9] == 0xff &&
		m.b[10] == 0xff &&
		m.b[11] == 0xff &&
		m.b[12] == 0xff &&
		m.b[13] == 0xff &&
		m.b[14] == 0xff &&
		m.b[15] == 0xff

}

// Inc ...
func (m *Marker) inc() *Marker {
	if isFirst(m) || isLast(m) {
		return m
	}

	b := make([]byte, 16)
	copy(b[:8], m.b[:8])
	binary.BigEndian.PutUint64(b[8:],
		binary.BigEndian.Uint64(m.b[8:])+1)
	return &Marker{b}
}

func (m *Marker) time() time.Time {
	return time.Unix(0,
		int64(binary.BigEndian.Uint64(m.b[8:])))
}

func (m *Marker) ip() net.IP {
	return net.IP(m.b[:8])
}

// NewMarker ...
func NewMarker(ip net.IP, t time.Time) *Marker {
	b := make([]byte, 16)
	copy(b[:8], ip)
	binary.BigEndian.PutUint64(b[8:], uint64(t.UnixNano()))
	return &Marker{b}
}

func marshalResults(r []time.Duration) []byte {
	b := bytes.NewBuffer(make([]byte, 0, len(r)*8))
	for _, t := range r {
		binary.Write(b, binary.BigEndian, t.Nanoseconds())
	}
	return b.Bytes()
}

func unmarshalResults(b []byte) []time.Duration {
	n := len(b) / 8
	d := make([]time.Duration, n)
	for i := 0; i < n; i++ {
		d[i] = time.Duration(
			binary.BigEndian.Uint64(b[i*8 : (i+1)*8]))
	}
	return d
}

// Write ...
func (s *Store) Write(ip net.IP, t time.Time, r []time.Duration) error {
	return s.db.Put(
		NewMarker(ip, t).b,
		marshalResults(r),
		nil)
}

func newIterator(fr, to *Marker) (*util.Range, bool) {
	c := bytes.Compare(fr.b, to.b)
	if c > 0 {
		return &util.Range{
			Start: to.b,
			Limit: fr.inc().b,
		}, false
	}

	return &util.Range{
		Start: fr.b,
		Limit: to.inc().b,
	}, true
}

func disp(
	it iterator.Iterator,
	fn func(ip net.IP, t time.Time, r []time.Duration) error) error {
	m := Marker{
		b: it.Key(),
	}
	return fn(m.ip(), m.time(), unmarshalResults(it.Value()))
}

// ForEach ...
func (s *Store) ForEach(
	fr, to *Marker,
	fn func(ip net.IP, t time.Time, r []time.Duration) error) error {
	r, fwd := newIterator(fr, to)
	if r == nil {
		return nil
	}

	it := s.db.NewIterator(r, nil)
	defer it.Release()

	if fwd {
		for it.Next() {
			if err := disp(it, fn); err == ErrStop {
				return nil
			} else if err != nil {
				return err
			}
		}
	} else {
		if it.Last() {
			if err := disp(it, fn); err == ErrStop {
				return nil
			} else if err != nil {
				return err
			}
		}
		for it.Prev() {
			if err := disp(it, fn); err == ErrStop {
				return nil
			} else if err != nil {
				return err
			}
		}
	}

	return it.Error()
}

// Open ...
func Open(filename string) (*Store, error) {
	db, err := leveldb.OpenFile(filename, nil)
	if err != nil {
		return nil, errors.Wrap(err, "leveldb open")
	}

	return &Store{
		db: db,
	}, nil
}
