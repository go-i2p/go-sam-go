package sam3

import (
	"context"
	"net"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/go-sam-go/stream"
	"github.com/go-i2p/i2pkeys"
)

// StreamSession represents a streaming session
type StreamSession struct {
	Timeout       time.Duration
	Deadline      time.Time
	session       *stream.StreamSession
	sam           *common.SAM
	fromPort      string
	toPort        string
	signatureType string
}

// NewStreamSession creates a new StreamSession
func (sam *SAM) NewStreamSession(id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error) {
	session, err := stream.NewStreamSession(sam.sam, id, keys, options)
	if err != nil {
		return nil, err
	}

	return &StreamSession{
		session:       session,
		sam:           sam.sam,
		fromPort:      "",
		toPort:        "",
		signatureType: common.SIG_EdDSA_SHA512_Ed25519,
	}, nil
}

// NewStreamSessionWithSignature creates a new StreamSession with custom signature
func (sam *SAM) NewStreamSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error) {
	streamSAM := &stream.SAM{SAM: sam.sam}
	session, err := streamSAM.NewStreamSessionWithSignature(id, keys, options, sigType)
	if err != nil {
		return nil, err
	}

	return &StreamSession{
		session:       session,
		sam:           sam.sam,
		fromPort:      "",
		toPort:        "",
		signatureType: sigType,
	}, nil
}

// NewStreamSessionWithSignatureAndPorts creates a new StreamSession with signature and ports
func (sam *SAM) NewStreamSessionWithSignatureAndPorts(id, from, to string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error) {
	streamSAM := &stream.SAM{SAM: sam.sam}
	session, err := streamSAM.NewStreamSessionWithPorts(id, from, to, keys, options)
	if err != nil {
		return nil, err
	}

	return &StreamSession{
		session:       session,
		sam:           sam.sam,
		fromPort:      from,
		toPort:        to,
		signatureType: sigType,
	}, nil
}

// ID returns the local tunnel name
func (s *StreamSession) ID() string {
	return s.session.ID()
}

// Keys returns the keys associated with the session
func (s *StreamSession) Keys() i2pkeys.I2PKeys {
	return s.session.Keys()
}

// Addr returns the I2P destination address
func (s *StreamSession) Addr() i2pkeys.I2PAddr {
	return s.session.Keys().Addr()
}

// Close closes the session
func (s *StreamSession) Close() error {
	return s.session.Close()
}

// Listen creates a new stream listener
func (s *StreamSession) Listen() (*StreamListener, error) {
	listener, err := s.session.Listen()
	if err != nil {
		return nil, err
	}

	return &StreamListener{
		listener: listener,
	}, nil
}

// Dial establishes a connection to an address
func (s *StreamSession) Dial(n, addr string) (c net.Conn, err error) {
	dialer := s.session.NewDialer()
	return dialer.Dial(n, addr)
}

// DialContext establishes a connection with context
func (s *StreamSession) DialContext(ctx context.Context, n, addr string) (net.Conn, error) {
	dialer := s.session.NewDialer()
	return dialer.DialContext(ctx, n, addr)
}

// DialContextI2P establishes an I2P connection with context
func (s *StreamSession) DialContextI2P(ctx context.Context, n, addr string) (*SAMConn, error) {
	dialer := s.session.NewDialer()
	conn, err := dialer.DialContext(ctx, n, addr)
	if err != nil {
		return nil, err
	}
	return &SAMConn{conn: conn}, nil
}

// DialI2P dials to an I2P destination
func (s *StreamSession) DialI2P(addr i2pkeys.I2PAddr) (*SAMConn, error) {
	dialer := s.session.NewDialer()
	conn, err := dialer.DialI2P(addr)
	if err != nil {
		return nil, err
	}
	return &SAMConn{conn: conn}, nil
}

// Lookup performs name lookup
func (s *StreamSession) Lookup(name string) (i2pkeys.I2PAddr, error) {
	return s.sam.Lookup(name)
}

// Read reads data from the stream
func (s *StreamSession) Read(buf []byte) (int, error) {
	return s.session.Read(buf)
}

// Write sends data over the stream
func (s *StreamSession) Write(data []byte) (int, error) {
	return s.session.Write(data)
}

// LocalAddr returns the local address
func (s *StreamSession) LocalAddr() net.Addr {
	return s.session.LocalAddr()
}

// RemoteAddr returns the remote address
func (s *StreamSession) RemoteAddr() net.Addr {
	return s.session.RemoteAddr()
}

// SetDeadline sets read and write deadlines
func (s *StreamSession) SetDeadline(t time.Time) error {
	return s.session.SetDeadline(t)
}

// SetReadDeadline sets read deadline
func (s *StreamSession) SetReadDeadline(t time.Time) error {
	return s.session.SetReadDeadline(t)
}

// SetWriteDeadline sets write deadline
func (s *StreamSession) SetWriteDeadline(t time.Time) error {
	return s.session.SetWriteDeadline(t)
}

// From returns the from port
func (s *StreamSession) From() string {
	return s.fromPort
}

// To returns the to port
func (s *StreamSession) To() string {
	return s.toPort
}

// SignatureType returns the signature type
func (s *StreamSession) SignatureType() string {
	return s.signatureType
}
