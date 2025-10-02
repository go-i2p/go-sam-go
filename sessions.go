package sam

import (
	"github.com/go-i2p/go-sam-go/primary"
	"github.com/go-i2p/go-sam-go/raw"
	"github.com/go-i2p/go-sam-go/stream"
	"github.com/go-i2p/i2pkeys"
)

// NewPrimarySession creates a new primary session that can manage multiple sub-sessions
// of different types (stream, datagram, raw) within a single I2P session context.
// The primary session enables complex applications with multiple communication patterns
// while sharing the same I2P identity and tunnel infrastructure for enhanced efficiency.
//
// The session ID must be unique and will be used to identify this session in the SAM
// protocol. The I2P keys define the cryptographic identity that will be shared across
// all sub-sessions created from this primary session. Configuration options control
// tunnel parameters such as length, backup quantity, and other I2P-specific settings.
//
// Returns a PrimarySession that can create and manage sub-sessions, or an error if
// the session creation fails due to SAM protocol errors, network issues, or invalid
// configuration parameters.
//
// Example usage:
//
//	sam, err := NewSAM("127.0.0.1:7656")
//	keys, _ := i2pkeys.NewKeys(i2pkeys.KT_ECDSA_SHA256_P256)
//	primary, err := sam.NewPrimarySession("app-primary", keys, Options_Default)
//	streamSub, err := primary.NewStreamSubSession("tcp-handler", []string{})
//	datagramSub, err := primary.NewDatagramSubSession("udp-handler", []string{})
func (sam *SAM) NewPrimarySession(id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error) {
	return primary.NewPrimarySession(sam.SAM, id, keys, options)
}

// NewPrimarySessionWithSignature creates a new primary session with a specific signature type.
// This method provides the same functionality as NewPrimarySession but allows explicit
// control over the cryptographic signature algorithm used for the session's I2P identity.
//
// The signature type must be one of the supported I2P signature algorithms (use the Sig_*
// constants defined in this package). Different signature types offer different security
// levels and performance characteristics - EdDSA_SHA512_Ed25519 is recommended for most
// applications as it provides strong security with good performance.
//
// Returns a PrimarySession configured with the specified signature type, or an error if
// the session creation fails or the signature type is not supported by the I2P router.
//
// Example usage:
//
//	sam, err := NewSAM("127.0.0.1:7656")
//	keys, _ := i2pkeys.NewKeys(i2pkeys.KT_EdDSA_SHA512_Ed25519)
//	primary, err := sam.NewPrimarySessionWithSignature("secure-primary", keys,
//		Options_Default, Sig_EdDSA_SHA512_Ed25519)
func (sam *SAM) NewPrimarySessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error) {
	return primary.NewPrimarySessionWithSignature(sam.SAM, id, keys, options, sigType)
}

// NewStreamSession creates a new stream session for TCP-like reliable connections over I2P.
// Stream sessions provide connection-oriented communication with guarantees for message
// ordering, delivery, and flow control, similar to TCP but routed through I2P tunnels.
//
// The session can be used to create listeners for accepting incoming connections or to
// establish outbound connections to other I2P destinations. All connections share the
// same I2P identity defined by the provided keys and benefit from the tunnel configuration
// specified in the options.
//
// Returns a StreamSession ready for creating connections, or an error if the session
// creation fails due to SAM protocol errors, network issues, or invalid configuration.
//
// Example usage:
//
//	sam, err := NewSAM("127.0.0.1:7656")
//	keys, _ := i2pkeys.NewKeys(i2pkeys.KT_ECDSA_SHA256_P256)
//	session, err := sam.NewStreamSession("web-server", keys, Options_Default)
//	listener, err := session.Listen()
//	conn, err := session.Dial("destination.b32.i2p")
func (sam *SAM) NewStreamSession(id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error) {
	return stream.NewStreamSession(sam.SAM, id, keys, options)
}

// NewStreamSessionWithSignature creates a new stream session with a specific signature type.
// This method provides the same functionality as NewStreamSession but allows explicit
// control over the cryptographic signature algorithm used for the session's I2P identity.
//
// The signature type determines the cryptographic strength and performance characteristics
// of the session. EdDSA_SHA512_Ed25519 is recommended for most applications, while
// ECDSA variants may be preferred for compatibility with older I2P routers or specific
// security requirements.
//
// Returns a StreamSession configured with the specified signature type, or an error if
// the session creation fails or the signature type is not supported.
//
// Example usage:
//
//	sam, err := NewSAM("127.0.0.1:7656")
//	keys, _ := i2pkeys.NewKeys(i2pkeys.KT_EdDSA_SHA512_Ed25519)
//	session, err := sam.NewStreamSessionWithSignature("secure-stream", keys,
//		Options_Large, Sig_EdDSA_SHA512_Ed25519)
func (sam *SAM) NewStreamSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error) {
	return stream.NewStreamSessionWithSignature(sam.SAM, id, keys, options, sigType)
}

// NewStreamSessionWithSignatureAndPorts creates a new stream session with signature type
// and specific port configuration. This method enables advanced port mapping scenarios
// where the session should bind to specific local ports or forward to specific remote ports.
//
// The 'from' parameter specifies the local port or port range that the session should
// use for incoming connections. The 'to' parameter specifies the target port or port
// range for outbound connections. Port specifications can be single ports ("80") or
// ranges ("8080-8090").
//
// This method is particularly useful for applications that need to maintain consistent
// port mappings or integrate with existing network infrastructure that expects specific
// port configurations.
//
// Returns a StreamSession with the specified port configuration, or an error if the
// session creation fails or the port configuration is invalid.
//
// Example usage:
//
//	sam, err := NewSAM("127.0.0.1:7656")
//	keys, _ := i2pkeys.NewKeys(i2pkeys.KT_ECDSA_SHA256_P256)
//	session, err := sam.NewStreamSessionWithSignatureAndPorts("http-proxy",
//		"8080", "80", keys, Options_Default, Sig_ECDSA_SHA256_P256)
/*func (sam *SAM) NewStreamSessionWithSignatureAndPorts(id, from, to string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error) {
	return stream.NewStreamSessionWithSignatureAndPorts(sam.SAM, id, from, to, keys, options, sigType)
}*/

// NewDatagramSession creates a new datagram session for UDP-like authenticated messaging over I2P.
// Datagram sessions provide connectionless communication with message authentication and
// replay protection, suitable for applications that need fast, lightweight messaging
// without the overhead of connection establishment.
//
// Unlike raw sessions, datagram sessions provide sender authentication and message
// integrity verification, making them suitable for applications where message authenticity
// is important. Each message includes cryptographic signatures that allow recipients to
// verify the sender's identity and detect tampering.
//
// The udpPort parameter specifies the local UDP port that the session should use for
// sending and receiving datagrams. This port is used for the local UDP socket that
// interfaces with the I2P datagram subsystem.
//
// Returns a DatagramSession ready for sending and receiving authenticated datagrams,
// or an error if the session creation fails.
//
// Example usage:
//
//	sam, err := NewSAM("127.0.0.1:7656")
//	keys, _ := i2pkeys.NewKeys(i2pkeys.KT_ECDSA_SHA256_P256)
//	session, err := sam.NewDatagramSession("chat-app", keys, Options_Medium, 0)
//	writer := session.NewWriter()
//	reader := session.NewReader()
/*func (sam *SAM) NewDatagramSession(id string, keys i2pkeys.I2PKeys, options []string, udpPort int) (*DatagramSession, error) {
	return datagram.NewDatagramSession(sam.SAM, id, keys, options)
}*/

// NewRawSession creates a new raw session for unrepliable datagram communication over I2P.
// Raw sessions provide the most lightweight form of I2P communication, where messages
// are encrypted end-to-end but do not include sender authentication or guarantee delivery.
// Recipients cannot identify the sender or send replies directly.
//
// Raw sessions are suitable for applications that need maximum anonymity and minimal
// overhead, such as anonymous publishing, voting systems, or situations where sender
// identity must remain completely hidden. The lack of sender authentication provides
// stronger anonymity but eliminates the ability to verify message authenticity.
//
// The udpPort parameter specifies the local UDP port for the session's datagram interface.
// This port is used for the local UDP socket that communicates with the I2P raw datagram
// subsystem.
//
// Returns a RawSession ready for sending anonymous unrepliable datagrams, or an error
// if the session creation fails.
//
// Example usage:
//
//	sam, err := NewSAM("127.0.0.1:7656")
//	keys, _ := i2pkeys.NewKeys(i2pkeys.KT_ECDSA_SHA256_P256)
//	session, err := sam.NewRawSession("anonymous-publisher", keys, Options_Small, 0)
//	writer := session.NewWriter()
//	reader := session.NewReader()
func (sam *SAM) NewRawSession(id string, keys i2pkeys.I2PKeys, options []string, udpPort int) (*RawSession, error) {
	return raw.NewRawSession(sam.SAM, id, keys, options)
}
