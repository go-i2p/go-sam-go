package sam3

import (
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/go-sam-go/raw"
	"github.com/go-i2p/i2pkeys"
)

// RawSession provides raw datagram messaging
type RawSession struct {
	session *raw.RawSession
	sam     *common.SAM
}

// NewRawSession creates a new raw session
func (s *SAM) NewRawSession(id string, keys i2pkeys.I2PKeys, options []string, udpPort int) (*RawSession, error) {
	session, err := raw.NewRawSession(s.sam, id, keys, options)
	if err != nil {
		return nil, err
	}

	return &RawSession{
		session: session,
		sam:     s.sam,
	}, nil
}

// Read reads one raw datagram
func (s *RawSession) Read(b []byte) (n int, err error) {
	return s.session.Read(b)
}

// WriteTo sends one raw datagram to the destination
func (s *RawSession) WriteTo(b []byte, addr i2pkeys.I2PAddr) (n int, err error) {
	return s.session.WriteTo(b, addr)
}

// Close closes the raw session
func (s *RawSession) Close() error {
	return s.session.Close()
}

// LocalAddr returns the local I2P destination
func (s *RawSession) LocalAddr() i2pkeys.I2PAddr {
	return s.session.Addr()
}

// SetDeadline sets read and write deadlines
func (s *RawSession) SetDeadline(t time.Time) error {
	return s.session.SetDeadline(t)
}

// SetReadDeadline sets read deadline
func (s *RawSession) SetReadDeadline(t time.Time) error {
	return s.session.SetReadDeadline(t)
}

// SetWriteDeadline sets write deadline
func (s *RawSession) SetWriteDeadline(t time.Time) error {
	return s.session.SetWriteDeadline(t)
}
