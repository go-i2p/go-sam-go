// Package sam3 provides configuration wrapper functions for exposing common package
// configuration functionality at the root level. This file implements wrapper functions
// for SAMEmit options and I2PConfig creation that enable applications to configure
// I2P tunnel parameters and session behaviors through a clean, sam3-compatible API.
package sam3

import (
	"github.com/go-i2p/go-sam-go/common"
)

// Configuration creation functions - These provide simplified constructors for
// I2PConfig and SAMEmit instances with functional options pattern support.

// NewConfig creates a new I2PConfig instance with default values and applies functional options.
// Returns a configured instance ready for use in session creation or an error if any option fails.
// This delegates to common.NewConfig while providing the sam3 API surface.
//
// Example usage:
//
//	config, err := NewConfig(SetInLength(4), SetOutLength(4))
//	if err != nil {
//	    log.Fatal("Failed to create config:", err)
//	}
func NewConfig(opts ...func(*I2PConfig) error) (*I2PConfig, error) {
	return common.NewConfig(opts...)
}

// NewEmit creates a new SAMEmit instance with the specified configuration options.
// Applies functional options to configure the emitter with custom settings.
// Returns an error if any option fails to apply correctly.
//
// Example usage:
//
//	emit, err := NewEmit(SetSAMHost("localhost"), SetSAMPort("7656"))
//	if err != nil {
//	    log.Fatal("Failed to create emitter:", err)
//	}
func NewEmit(opts ...func(*SAMEmit) error) (*SAMEmit, error) {
	return common.NewEmit(opts...)
}

// SAMEmit configuration functions - These provide functional options for configuring
// SAMEmit instances with I2P tunnel parameters, SAM bridge settings, and session behaviors.

// SetType sets the session type for the forwarder server.
// Valid values are "STREAM", "DATAGRAM", or "RAW" corresponding to different
// communication patterns available in the I2P network protocol.
//
// Example usage:
//
//	emit, err := NewEmit(SetType("STREAM"))
func SetType(s string) func(*SAMEmit) error {
	return common.SetType(s)
}

// SetSAMAddress sets the SAM bridge address all-at-once using "host:port" format.
// This convenience function parses the address string and sets both host and port
// simultaneously for simplified SAM bridge configuration.
//
// Example usage:
//
//	emit, err := NewEmit(SetSAMAddress("127.0.0.1:7656"))
func SetSAMAddress(s string) func(*SAMEmit) error {
	return common.SetSAMAddress(s)
}

// SetSAMHost sets the hostname or IP address of the SAM bridge.
// The SAM bridge is the interface that allows applications to communicate
// with the I2P router for creating sessions and managing connections.
//
// Example usage:
//
//	emit, err := NewEmit(SetSAMHost("127.0.0.1"))
func SetSAMHost(s string) func(*SAMEmit) error {
	return common.SetSAMHost(s)
}

// SetSAMPort sets the port number of the SAM bridge using a string value.
// The port must be a valid TCP port number (0-65535) where the I2P router's
// SAM bridge is listening for incoming connections.
//
// Example usage:
//
//	emit, err := NewEmit(SetSAMPort("7656"))
func SetSAMPort(s string) func(*SAMEmit) error {
	return common.SetSAMPort(s)
}

// SetName sets the tunnel name for identification and debugging purposes.
// This name appears in I2P router logs and management interfaces to help
// identify and manage specific tunnels created by your application.
//
// Example usage:
//
//	emit, err := NewEmit(SetName("my-app-tunnel"))
func SetName(s string) func(*SAMEmit) error {
	return common.SetName(s)
}

// Tunnel length configuration functions - These control the number of hops
// in I2P tunnels, directly affecting anonymity levels and latency characteristics.

// SetInLength sets the number of hops for inbound tunnels (0-6).
// Higher values provide better anonymity but increase latency and resource usage.
// Most applications use 3 hops as a balance between security and performance.
//
// Example usage:
//
//	emit, err := NewEmit(SetInLength(3))
func SetInLength(u int) func(*SAMEmit) error {
	return common.SetInLength(u)
}

// SetOutLength sets the number of hops for outbound tunnels (0-6).
// Higher values provide better anonymity but increase latency and resource usage.
// Most applications use 3 hops as a balance between security and performance.
//
// Example usage:
//
//	emit, err := NewEmit(SetOutLength(3))
func SetOutLength(u int) func(*SAMEmit) error {
	return common.SetOutLength(u)
}

// SetInVariance sets the variance for inbound tunnel hop counts (-6 to 6).
// This adds randomness to tunnel lengths to prevent traffic analysis.
// Positive values increase maximum hops, negative values allow shorter tunnels.
//
// Example usage:
//
//	emit, err := NewEmit(SetInVariance(1))
func SetInVariance(i int) func(*SAMEmit) error {
	return common.SetInVariance(i)
}

// SetOutVariance sets the variance for outbound tunnel hop counts (-6 to 6).
// This adds randomness to tunnel lengths to prevent traffic analysis.
// Positive values increase maximum hops, negative values allow shorter tunnels.
//
// Example usage:
//
//	emit, err := NewEmit(SetOutVariance(1))
func SetOutVariance(i int) func(*SAMEmit) error {
	return common.SetOutVariance(i)
}

// Tunnel quantity configuration functions - These control the number of parallel
// tunnels maintained for load balancing and redundancy in I2P communications.

// SetInQuantity sets the number of inbound tunnels to maintain (1-16).
// More tunnels provide better load distribution and redundancy but consume
// more router resources. Most applications use 1-4 tunnels.
//
// Example usage:
//
//	emit, err := NewEmit(SetInQuantity(2))
func SetInQuantity(u int) func(*SAMEmit) error {
	return common.SetInQuantity(u)
}

// SetOutQuantity sets the number of outbound tunnels to maintain (1-16).
// More tunnels provide better load distribution and redundancy but consume
// more router resources. Most applications use 1-4 tunnels.
//
// Example usage:
//
//	emit, err := NewEmit(SetOutQuantity(2))
func SetOutQuantity(u int) func(*SAMEmit) error {
	return common.SetOutQuantity(u)
}

// SetInBackups sets the number of backup inbound tunnels (0-5).
// Backup tunnels are pre-built spare tunnels that can quickly replace
// failed primary tunnels, improving connection reliability.
//
// Example usage:
//
//	emit, err := NewEmit(SetInBackups(1))
func SetInBackups(u int) func(*SAMEmit) error {
	return common.SetInBackups(u)
}

// SetOutBackups sets the number of backup outbound tunnels (0-5).
// Backup tunnels are pre-built spare tunnels that can quickly replace
// failed primary tunnels, improving connection reliability.
//
// Example usage:
//
//	emit, err := NewEmit(SetOutBackups(1))
func SetOutBackups(u int) func(*SAMEmit) error {
	return common.SetOutBackups(u)
}

// Security and encryption configuration functions - These control cryptographic
// settings and lease set encryption for enhanced security in I2P communications.

// SetEncrypt enables or disables encrypted lease sets for enhanced security.
// Encrypted lease sets provide additional protection against traffic analysis
// but may slightly impact performance and compatibility.
//
// Example usage:
//
//	emit, err := NewEmit(SetEncrypt(true))
func SetEncrypt(b bool) func(*SAMEmit) error {
	return common.SetEncrypt(b)
}

// SetLeaseSetKey sets the public key for lease set encryption.
// This key is used to encrypt the lease set information, providing
// additional security for the destination's routing information.
//
// Example usage:
//
//	emit, err := NewEmit(SetLeaseSetKey("base64-encoded-key"))
func SetLeaseSetKey(s string) func(*SAMEmit) error {
	return common.SetLeaseSetKey(s)
}

// SetLeaseSetPrivateKey sets the private key for lease set decryption.
// This key is used to decrypt lease set information when encrypted
// lease sets are enabled for enhanced security.
//
// Example usage:
//
//	emit, err := NewEmit(SetLeaseSetPrivateKey("base64-encoded-private-key"))
func SetLeaseSetPrivateKey(s string) func(*SAMEmit) error {
	return common.SetLeaseSetPrivateKey(s)
}

// SetLeaseSetPrivateSigningKey sets the private signing key for lease set authentication.
// This key is used to sign lease set information to ensure authenticity
// and prevent tampering with routing information.
//
// Example usage:
//
//	emit, err := NewEmit(SetLeaseSetPrivateSigningKey("base64-encoded-signing-key"))
func SetLeaseSetPrivateSigningKey(s string) func(*SAMEmit) error {
	return common.SetLeaseSetPrivateSigningKey(s)
}

// Protocol and reliability configuration functions - These control message
// reliability, compression, and protocol optimization settings.

// SetMessageReliability sets the reliability level for message delivery.
// Options include "none", "BestEffort", or "Guaranteed" depending on
// the application's reliability requirements and performance trade-offs.
//
// Example usage:
//
//	emit, err := NewEmit(SetMessageReliability("BestEffort"))
func SetMessageReliability(s string) func(*SAMEmit) error {
	return common.SetMessageReliability(s)
}

// SetAllowZeroIn enables or disables acceptance of zero-hop inbound tunnels.
// Zero-hop tunnels provide no anonymity but offer better performance for
// applications that don't require anonymity protection.
//
// Example usage:
//
//	emit, err := NewEmit(SetAllowZeroIn(false))
func SetAllowZeroIn(b bool) func(*SAMEmit) error {
	return common.SetAllowZeroIn(b)
}

// SetAllowZeroOut enables or disables acceptance of zero-hop outbound tunnels.
// Zero-hop tunnels provide no anonymity but offer better performance for
// applications that don't require anonymity protection.
//
// Example usage:
//
//	emit, err := NewEmit(SetAllowZeroOut(false))
func SetAllowZeroOut(b bool) func(*SAMEmit) error {
	return common.SetAllowZeroOut(b)
}

// SetCompress enables or disables data compression for tunnel traffic.
// Compression can reduce bandwidth usage but may impact performance and
// could potentially affect anonymity through traffic analysis.
//
// Example usage:
//
//	emit, err := NewEmit(SetCompress(true))
func SetCompress(b bool) func(*SAMEmit) error {
	return common.SetCompress(b)
}

// SetFastRecieve enables or disables fast receive mode for improved performance.
// When enabled, bypasses some protocol overhead for faster data transmission
// at the potential cost of some reliability guarantees.
//
// Example usage:
//
//	emit, err := NewEmit(SetFastRecieve(true))
func SetFastRecieve(b bool) func(*SAMEmit) error {
	return common.SetFastRecieve(b)
}

// Idle management configuration functions - These control tunnel behavior
// during periods of inactivity to optimize resource usage.

// SetReduceIdle enables or disables tunnel reduction during extended idle periods.
// When enabled, reduces the number of active tunnels during idle time to
// conserve router resources while maintaining minimal connectivity.
//
// Example usage:
//
//	emit, err := NewEmit(SetReduceIdle(true))
func SetReduceIdle(b bool) func(*SAMEmit) error {
	return common.SetReduceIdle(b)
}

// SetReduceIdleTime sets the time in minutes to wait before reducing tunnels.
// After this period of inactivity, the number of tunnels will be reduced
// to the quantity specified by SetReduceIdleQuantity.
//
// Example usage:
//
//	emit, err := NewEmit(SetReduceIdleTime(10)) // 10 minutes
func SetReduceIdleTime(u int) func(*SAMEmit) error {
	return common.SetReduceIdleTime(u)
}

// SetReduceIdleTimeMs sets the time in milliseconds to wait before reducing tunnels.
// After this period of inactivity, the number of tunnels will be reduced
// to the quantity specified by SetReduceIdleQuantity.
//
// Example usage:
//
//	emit, err := NewEmit(SetReduceIdleTimeMs(600000)) // 10 minutes
func SetReduceIdleTimeMs(u int) func(*SAMEmit) error {
	return common.SetReduceIdleTimeMs(u)
}

// SetReduceIdleQuantity sets the minimum number of tunnels during idle periods.
// When idle reduction is enabled, this is the number of tunnels that will
// be maintained during periods of low activity.
//
// Example usage:
//
//	emit, err := NewEmit(SetReduceIdleQuantity(1))
func SetReduceIdleQuantity(u int) func(*SAMEmit) error {
	return common.SetReduceIdleQuantity(u)
}

// SetCloseIdle enables or disables complete tunnel closure during extended idle periods.
// When enabled, closes all tunnels after the specified idle time to
// completely conserve router resources.
//
// Example usage:
//
//	emit, err := NewEmit(SetCloseIdle(true))
func SetCloseIdle(b bool) func(*SAMEmit) error {
	return common.SetCloseIdle(b)
}

// SetCloseIdleTime sets the time in minutes to wait before closing all tunnels.
// After this period of inactivity, all tunnels will be closed to
// conserve router resources.
//
// Example usage:
//
//	emit, err := NewEmit(SetCloseIdleTime(30)) // 30 minutes
func SetCloseIdleTime(u int) func(*SAMEmit) error {
	return common.SetCloseIdleTime(u)
}

// SetCloseIdleTimeMs sets the time in milliseconds to wait before closing all tunnels.
// After this period of inactivity, all tunnels will be closed to
// conserve router resources.
//
// Example usage:
//
//	emit, err := NewEmit(SetCloseIdleTimeMs(1800000)) // 30 minutes
func SetCloseIdleTimeMs(u int) func(*SAMEmit) error {
	return common.SetCloseIdleTimeMs(u)
}

// Access control configuration functions - These manage connection filtering
// through whitelist and blacklist mechanisms for destination-based access control.

// SetAccessListType sets the type of access control to apply.
// Valid values are "whitelist" (allow only listed destinations),
// "blacklist" (deny listed destinations), or "none" (no filtering).
//
// Example usage:
//
//	emit, err := NewEmit(SetAccessListType("whitelist"))
func SetAccessListType(s string) func(*SAMEmit) error {
	return common.SetAccessListType(s)
}

// SetAccessList sets the list of destinations for access control.
// The behavior depends on the access list type: whitelist allows only
// these destinations, blacklist denies these destinations.
//
// Example usage:
//
//	destinations := []string{"dest1.b32.i2p", "dest2.b32.i2p"}
//	emit, err := NewEmit(SetAccessList(destinations))
func SetAccessList(s []string) func(*SAMEmit) error {
	return common.SetAccessList(s)
}
