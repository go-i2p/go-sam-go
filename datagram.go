package sam3

import (
	"net"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/go-sam-go/datagram"
	"github.com/go-i2p/i2pkeys"
)

// DatagramSession implements net.PacketConn for I2P datagrams
type DatagramSession struct {
	session *datagram.DatagramSession
	sam     *common.SAM
}

// NewDatagramSession creates a new datagram session
func (s *SAM) NewDatagramSession(id string, keys i2pkeys.I2PKeys, options []string, udpPort int) (*DatagramSession, error) {
	session, err := datagram.NewDatagramSession(s.sam, id, keys, options)
	if err != nil {
		return nil, err
	}

	return &DatagramSession{
		session: session,
		sam:     s.sam,
	}, nil
}

// ReadFrom reads a datagram from the session
func (s *DatagramSession) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	return s.session.ReadFrom(b)
}

// WriteTo writes a datagram to the specified address
func (s *DatagramSession) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return s.session.WriteTo(b, addr)
}

// Close closes the datagram session
func (s *DatagramSession) Close() error {
	return s.session.Close()
}

// LocalAddr returns the local address
func (s *DatagramSession) LocalAddr() net.Addr {
	return s.session.LocalAddr()
}

// LocalI2PAddr returns the I2P destination
func (s *DatagramSession) LocalI2PAddr() i2pkeys.I2PAddr {
	return s.session.Addr()
}

// SetDeadline sets read and write deadlines
func (s *DatagramSession) SetDeadline(t time.Time) error {
	return s.session.SetDeadline(t)
}

// SetReadDeadline sets read deadline
func (s *DatagramSession) SetReadDeadline(t time.Time) error {
	return s.session.SetReadDeadline(t)
}

// SetWriteDeadline sets write deadline
func (s *DatagramSession) SetWriteDeadline(t time.Time) error {
	return s.session.SetWriteDeadline(t)
}

// Read reads from the session
func (s *DatagramSession) Read(b []byte) (n int, err error) {
	return s.session.Read(b)
}

// Write writes to the session
func (s *DatagramSession) Write(b []byte) (int, error) {
	return s.session.Write(b)
}

// Addr returns the session address
func (s *DatagramSession) Addr() net.Addr {
	return s.session.LocalAddr()
}

// B32 returns the base32 address
func (s *DatagramSession) B32() string {
	return s.session.Addr().Base32()
}

// RemoteAddr returns the remote address
func (s *DatagramSession) RemoteAddr() net.Addr {
	return s.session.RemoteAddr()
}

// Lookup performs name lookup
func (s *DatagramSession) Lookup(name string) (a net.Addr, err error) {
	addr, err := s.sam.Lookup(name)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// Accept accepts connections (not applicable for datagrams)
func (s *DatagramSession) Accept() (net.Conn, error) {
	return nil, net.ErrClosed
}

// Dial dials a connection (returns self for datagrams)
func (s *DatagramSession) Dial(net string, addr string) (*DatagramSession, error) {
	return s, nil
}

// DialI2PRemote dials to I2P remote
func (s *DatagramSession) DialI2PRemote(net string, addr net.Addr) (*DatagramSession, error) {
	return s, nil
}

// DialRemote dials to remote address
func (s *DatagramSession) DialRemote(net, addr string) (net.PacketConn, error) {
	return s, nil
}

// SetWriteBuffer sets write buffer size
func (s *DatagramSession) SetWriteBuffer(bytes int) error {
	// Not implemented in underlying library
	return nil
}
