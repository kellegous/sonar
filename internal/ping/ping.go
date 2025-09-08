package ping

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	protocolICMP = 1
)

// ErrNoAnswer ...
var ErrNoAnswer = errors.New("no answer")

// Results ...
type Results struct {
	Data []time.Duration
}

func timeToBytes(b []byte, t time.Time) {
	binary.LittleEndian.PutUint64(b, uint64(t.UnixNano()))
}

func bytesToTime(b []byte) time.Time {
	return time.Unix(0, int64(binary.LittleEndian.Uint64(b)))
}

func decodeTimeFrom(msg *icmp.Message) (int, int, time.Time, error) {
	b, err := msg.Body.Marshal(protocolICMP)
	if err != nil {
		return 0, 0, time.Time{}, err
	} else if len(b) != 12 {
		return 0, 0, time.Time{}, fmt.Errorf("unexpected length: %d", len(b))
	}

	id := int(binary.BigEndian.Uint16(b[:2]))
	seq := int(binary.BigEndian.Uint16(b[2:4]))

	return id, seq, bytesToTime(b[4:]), nil
}

func recv(p *Pinger, seq int, dl time.Time) (time.Duration, error) {
	if err := p.c.SetReadDeadline(dl); err != nil {
		return time.Duration(0), err
	}

	buf := p.b[:]

	n, _, err := p.c.ReadFrom(buf)
	if err != nil {
		return time.Duration(0), err
	}

	msg, err := icmp.ParseMessage(protocolICMP, buf[:n])
	if err != nil {
		return time.Duration(0), err
	}

	switch msg.Type {
	case ipv4.ICMPTypeEchoReply:
		id, s, t, err := decodeTimeFrom(msg)
		if err != nil {
			return time.Duration(0), err
		}

		// we received something, but not the right something.
		if p.id != id || seq != s {
			return recv(p, seq, dl)
		}

		return time.Since(t), nil
	case ipv4.ICMPTypeDestinationUnreachable, ipv4.ICMPTypeTimeExceeded:
		return time.Duration(0), ErrNoAnswer
	default:
		return time.Duration(0), fmt.Errorf("unexpected type: %v", msg.Type)
	}
}

// Pinger ...
type Pinger struct {
	c  *icmp.PacketConn
	id int
	b  [1500]byte
}

// Close ...
func (p *Pinger) Close() error {
	return p.c.Close()
}

// NewPinger ...
func NewPinger(id int) (*Pinger, error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}

	return &Pinger{
		c:  c,
		id: id,
	}, nil
}

// Ping ...
func (p *Pinger) Ping(ip net.IP, seq int) (time.Duration, error) {

	var data [8]byte

	timeToBytes(data[:], time.Now())

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.id & 0xffff,
			Seq:  seq,
			Data: data[:],
		},
	}

	mb, err := msg.Marshal(nil)
	if err != nil {
		return time.Duration(0), err
	}

	if _, err := p.c.WriteTo(mb, &net.IPAddr{IP: ip}); err != nil {
		return time.Duration(0), err
	}

	return recv(p, seq, time.Now().Add(5*time.Second))
}
