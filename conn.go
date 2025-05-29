package sam3

import (
	"net"
	"time"
)

// SAMConn implements net.Conn for I2P connections
type SAMConn struct {
	conn net.Conn
}

// Read reads data from the connection
func (sc *SAMConn) Read(buf []byte) (int, error) {
	return sc.conn.Read(buf)
}

// Write writes data to the connection
func (sc *SAMConn) Write(buf []byte) (int, error) {
	return sc.conn.Write(buf)
}

// Close closes the connection
func (sc *SAMConn) Close() error {
	return sc.conn.Close()
}

// LocalAddr returns the local address
func (sc *SAMConn) LocalAddr() net.Addr {
	return sc.conn.LocalAddr()
}

// RemoteAddr returns the remote address
func (sc *SAMConn) RemoteAddr() net.Addr {
	return sc.conn.RemoteAddr()
}

// SetDeadline sets read and write deadlines
func (sc *SAMConn) SetDeadline(t time.Time) error {
	return sc.conn.SetDeadline(t)
}

// SetReadDeadline sets read deadline
func (sc *SAMConn) SetReadDeadline(t time.Time) error {
	return sc.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets write deadline
func (sc *SAMConn) SetWriteDeadline(t time.Time) error {
	return sc.conn.SetWriteDeadline(t)
}
