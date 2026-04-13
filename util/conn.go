package util

import (
	"net"
	"time"
)

func NewDialerConn(conn net.Conn) (res *DialerConn) {
	res = &DialerConn{}
	res.conn = conn
	return
}

type DialerConn struct {
	conn net.Conn
}

func (t *DialerConn) Read(b []byte) (n int, err error) {
	return t.conn.Read(b)
}

func (t *DialerConn) Write(b []byte) (n int, err error) {
	return t.conn.Write(b)
}

func (t *DialerConn) Close() error {
	return t.conn.Close()
}

func (t *DialerConn) LocalAddr() net.Addr {
	return t.conn.LocalAddr()
}

func (t *DialerConn) RemoteAddr() net.Addr {
	return t.conn.RemoteAddr()
}

func (t *DialerConn) SetDeadline(deadline time.Time) error {
	if err := t.SetReadDeadline(deadline); err != nil {
		return err
	}
	return t.SetWriteDeadline(deadline)
}

func (t *DialerConn) SetReadDeadline(deadline time.Time) error {
	return nil
}

func (t *DialerConn) SetWriteDeadline(deadline time.Time) error {
	return nil
}
