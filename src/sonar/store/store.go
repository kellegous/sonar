package store

import (
	"bytes"
	"encoding/binary"
	"time"

	"sonar/config"
	"sonar/ping"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// Store ...
type Store struct {
	db *leveldb.DB
}

func marshalResults(r *ping.Results) []byte {
	b := bytes.NewBuffer(make([]byte, 0, len(r.Data)*8))
	for _, t := range r.Data {
		binary.Write(b, binary.BigEndian, t.Nanoseconds())
	}
	return b.Bytes()
}

func (s *Store) Write(h *config.Host, t time.Time, r *ping.Results) error {
	var key [16]byte
	copy(key[:8], h.IP)
	binary.BigEndian.PutUint64(key[8:], uint64(t.UnixNano()))

	val := marshalResults(r)

	return s.db.Put(key[:], val, nil)
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
