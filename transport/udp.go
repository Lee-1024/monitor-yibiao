package transport

import (
	"net"
	"time"
)

type Sender interface {
	Send([]byte) error
	Close() error
}
type UDP struct {
	conn    net.Conn
	timeout time.Duration
}

func DialUDP(addr string, timeout time.Duration) (*UDP, error) {
	c, e := net.Dial("udp", addr)
	if e != nil {
		return nil, e
	}
	return &UDP{c, timeout}, nil
}
func (u *UDP) Send(b []byte) error {
	if u.timeout > 0 {
		_ = u.conn.SetWriteDeadline(time.Now().Add(u.timeout))
	}
	_, e := u.conn.Write(b)
	return e
}
func (u *UDP) Close() error { return u.conn.Close() }
