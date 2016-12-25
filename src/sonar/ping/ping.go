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

func decodeTimeFrom(msg *icmp.Message) (time.Time, error) {
	b, err := msg.Body.Marshal(protocolICMP)
	if err != nil {
		return time.Time{}, err
	}

	return bytesToTime(b[4:]), nil
}

func recv(c *icmp.PacketConn, buf []byte) (time.Duration, error) {
	if err := c.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return time.Duration(0), err
	}

	n, _, err := c.ReadFrom(buf)
	if err != nil {
		return time.Duration(0), err
	}

	msg, err := icmp.ParseMessage(protocolICMP, buf[:n])
	if err != nil {
		return time.Duration(0), err
	}

	switch msg.Type {
	case ipv4.ICMPTypeEchoReply:
		t, err := decodeTimeFrom(msg)
		if err != nil {
			return time.Duration(0), err
		}
		return time.Now().Sub(t), nil
	case ipv4.ICMPTypeDestinationUnreachable, ipv4.ICMPTypeTimeExceeded:
		return time.Duration(0), ErrNoAnswer
	default:
		return time.Duration(0), fmt.Errorf("unexpected type: %v", msg.Type)
	}
}

// Pinger ...
type Pinger struct {
	c *icmp.PacketConn
	b [1500]byte
}

// Close ...
func (p *Pinger) Close() error {
	return p.c.Close()
}

// NewPinger ...
func NewPinger() (*Pinger, error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}

	return &Pinger{
		c: c,
	}, nil
}

// Ping ...
func (p *Pinger) Ping(ip net.IP, id int) (time.Duration, error) {

	var data [8]byte

	timeToBytes(data[:], time.Now())

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  id,
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

	return recv(p.c, p.b[:])
}
