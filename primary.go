package sam3

import (
	"net"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// PrimarySession represents a primary session
type PrimarySession struct {
	Timeout  time.Duration
	Deadline time.Time
	Config   SAMEmit
	sam      *common.SAM
	keys     i2pkeys.I2PKeys
	id       string
}

// NewPrimarySession creates a new PrimarySession
func (sam *SAM) NewPrimarySession(id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error) {
	return &PrimarySession{
		Config: sam.Config,
		sam:    sam.sam,
		keys:   keys,
		id:     id,
	}, nil
}

// NewPrimarySessionWithSignature creates a new PrimarySession with signature
func (sam *SAM) NewPrimarySessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error) {
	return &PrimarySession{
		Config: sam.Config,
		sam:    sam.sam,
		keys:   keys,
		id:     id,
	}, nil
}

// ID returns the session ID
func (ss *PrimarySession) ID() string {
	return ss.id
}

// Keys returns the session keys
func (ss *PrimarySession) Keys() i2pkeys.I2PKeys {
	return ss.keys
}

// Addr returns the I2P address
func (ss *PrimarySession) Addr() i2pkeys.I2PAddr {
	return ss.keys.Addr()
}

// Close closes the session
func (ss *PrimarySession) Close() error {
	return nil
}

// NewStreamSubSession creates a new stream sub-session
func (sam *PrimarySession) NewStreamSubSession(id string) (*StreamSession, error) {
	samWrapper := &SAM{sam: sam.sam}
	return samWrapper.NewStreamSession(id, sam.keys, nil)
}

// NewStreamSubSessionWithPorts creates a new stream sub-session with ports
func (sam *PrimarySession) NewStreamSubSessionWithPorts(id, from, to string) (*StreamSession, error) {
	samWrapper := &SAM{sam: sam.sam}
	return samWrapper.NewStreamSessionWithSignatureAndPorts(id, from, to, sam.keys, nil, Sig_EdDSA_SHA512_Ed25519)
}

// NewUniqueStreamSubSession creates a unique stream sub-session
func (sam *PrimarySession) NewUniqueStreamSubSession(id string) (*StreamSession, error) {
	return sam.NewStreamSubSession(id + RandString())
}

// NewDatagramSubSession creates a new datagram sub-session
func (s *PrimarySession) NewDatagramSubSession(id string, udpPort int) (*DatagramSession, error) {
	samWrapper := &SAM{sam: s.sam}
	return samWrapper.NewDatagramSession(id, s.keys, nil, udpPort)
}

// NewRawSubSession creates a new raw sub-session
func (s *PrimarySession) NewRawSubSession(id string, udpPort int) (*RawSession, error) {
	samWrapper := &SAM{sam: s.sam}
	return samWrapper.NewRawSession(id, s.keys, nil, udpPort)
}

// Dial implements net.Dialer
func (sam *PrimarySession) Dial(network, addr string) (net.Conn, error) {
	ss, err := sam.NewStreamSubSession("dial-" + RandString())
	if err != nil {
		return nil, err
	}
	return ss.Dial(network, addr)
}

// DialTCP implements x/dialer
func (sam *PrimarySession) DialTCP(network string, laddr, raddr net.Addr) (net.Conn, error) {
	return sam.Dial(network, raddr.String())
}

// DialTCPI2P dials TCP over I2P
func (sam *PrimarySession) DialTCPI2P(network string, laddr, raddr string) (net.Conn, error) {
	return sam.Dial(network, raddr)
}

// DialUDP implements x/dialer
func (sam *PrimarySession) DialUDP(network string, laddr, raddr net.Addr) (net.PacketConn, error) {
	ds, err := sam.NewDatagramSubSession("udp-"+RandString(), 0)
	if err != nil {
		return nil, err
	}
	return ds, nil
}

// DialUDPI2P dials UDP over I2P
func (sam *PrimarySession) DialUDPI2P(network, laddr, raddr string) (*DatagramSession, error) {
	return sam.NewDatagramSubSession("udp-"+RandString(), 0)
}

// Lookup performs name lookup
func (s *PrimarySession) Lookup(name string) (a net.Addr, err error) {
	addr, err := s.sam.Lookup(name)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// Resolve resolves network address
func (sam *PrimarySession) Resolve(network, addr string) (net.Addr, error) {
	return sam.Lookup(addr)
}

// ResolveTCPAddr resolves TCP address
func (sam *PrimarySession) ResolveTCPAddr(network, dest string) (net.Addr, error) {
	return sam.Lookup(dest)
}

// ResolveUDPAddr resolves UDP address
func (sam *PrimarySession) ResolveUDPAddr(network, dest string) (net.Addr, error) {
	return sam.Lookup(dest)
}

// LocalAddr returns local address
func (ss *PrimarySession) LocalAddr() net.Addr {
	return ss.keys.Addr()
}

// From returns from port
func (ss *PrimarySession) From() string {
	return ""
}

// To returns to port
func (ss *PrimarySession) To() string {
	return ""
}

// SignatureType returns signature type
func (ss *PrimarySession) SignatureType() string {
	return ""
}
