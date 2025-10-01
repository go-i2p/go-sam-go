// Package sam provides a pure-Go implementation of SAMv3.3 (Simple Anonymous Messaging) for I2P networks.
// This is the root package wrapper that provides sam3-compatible API surface while delegating
// implementation details to specialized sub-packages.
package sam

import (
	"os"
	"strings"
)

// Signature type constants for I2P destination key generation.
// These specify the cryptographic signature algorithm used for I2P destinations.
// SIG_DEFAULT points to the recommended secure signature type for new applications.
const (
	// Sig_NONE is deprecated, use Sig_EdDSA_SHA512_Ed25519 instead for secure signatures.
	Sig_NONE = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"

	// Sig_DSA_SHA1 specifies DSA with SHA1 signature type (legacy, not recommended for new applications).
	Sig_DSA_SHA1 = "SIGNATURE_TYPE=DSA_SHA1"

	// Sig_ECDSA_SHA256_P256 specifies ECDSA with SHA256 on P256 curve signature type.
	Sig_ECDSA_SHA256_P256 = "SIGNATURE_TYPE=ECDSA_SHA256_P256"

	// Sig_ECDSA_SHA384_P384 specifies ECDSA with SHA384 on P384 curve signature type.
	Sig_ECDSA_SHA384_P384 = "SIGNATURE_TYPE=ECDSA_SHA384_P384"

	// Sig_ECDSA_SHA512_P521 specifies ECDSA with SHA512 on P521 curve signature type.
	Sig_ECDSA_SHA512_P521 = "SIGNATURE_TYPE=ECDSA_SHA512_P521"

	// Sig_EdDSA_SHA512_Ed25519 specifies EdDSA with SHA512 on Ed25519 curve signature type (recommended).
	Sig_EdDSA_SHA512_Ed25519 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"

	// Sig_DEFAULT points to the recommended secure signature type for new applications.
	Sig_DEFAULT = Sig_EdDSA_SHA512_Ed25519
)

// Predefined tunnel configuration option sets for different traffic patterns and anonymity requirements.
// These option sets balance performance, anonymity, and resource usage for common use cases.
// Each option set specifies tunnel length, variance, backup quantity, and parallel tunnel count.
var (
	// Options_Humongous provides maximum anonymity and redundancy for extremely high-value traffic.
	// Suitable for applications requiring the highest level of anonymity protection with significant
	// resource overhead. Uses 3-hop tunnels with high redundancy and parallel connections.
	Options_Humongous = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=3", "outbound.backupQuantity=3",
		"inbound.quantity=6", "outbound.quantity=6",
	}

	// Options_Large provides strong anonymity for high-traffic applications.
	// Suitable for applications shuffling large amounts of traffic with good anonymity protection.
	// Balances performance and anonymity with reasonable resource usage.
	Options_Large = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=1", "outbound.backupQuantity=1",
		"inbound.quantity=4", "outbound.quantity=4",
	}

	// Options_Wide provides minimal anonymity but high performance for traffic requiring low latency.
	// Suitable for applications prioritizing speed over anonymity. Uses 1-hop tunnels with
	// moderate redundancy for basic privacy protection while maintaining performance.
	Options_Wide = []string{
		"inbound.length=1", "outbound.length=1",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=2", "outbound.backupQuantity=2",
		"inbound.quantity=3", "outbound.quantity=3",
	}

	// Options_Medium provides balanced anonymity and performance for moderate traffic loads.
	// Suitable for applications with medium traffic requirements that need good anonymity
	// without excessive resource overhead. Provides solid baseline protection.
	Options_Medium = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=2", "outbound.quantity=2",
	}

	// Options_Default provides sensible defaults for most use cases.
	// Recommended starting point for most applications. Provides good anonymity
	// with reasonable performance characteristics and resource usage.
	Options_Default = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=0", "outbound.lengthVariance=0",
		"inbound.backupQuantity=1", "outbound.backupQuantity=1",
		"inbound.quantity=1", "outbound.quantity=1",
	}

	// Options_Small provides basic anonymity for low-traffic, short-duration connections.
	// Suitable only for applications with minimal traffic requirements and short connection
	// lifetimes. Offers basic privacy with minimal resource usage.
	Options_Small = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=1", "outbound.quantity=1",
	}

	// Options_Warning_ZeroHop disables all anonymization - zero hop configuration.
	// WARNING: This configuration provides NO anonymity protection and should only be used
	// for testing or debugging purposes. All traffic is directly routed without tunnel protection.
	Options_Warning_ZeroHop = []string{
		"inbound.length=0", "outbound.length=0",
		"inbound.lengthVariance=0", "outbound.lengthVariance=0",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=2", "outbound.quantity=2",
	}
)

// SAM bridge connection variables with environment variable support.
// These variables can be overridden using sam_host and sam_port environment variables.
var (
	// SAM_HOST specifies the SAM bridge host address.
	// Can be overridden with the 'sam_host' environment variable.
	// Defaults to localhost (127.0.0.1) for local I2P router connections.
	SAM_HOST = getEnv("sam_host", "127.0.0.1")

	// SAM_PORT specifies the SAM bridge port number.
	// Can be overridden with the 'sam_port' environment variable.
	// Defaults to 7656, the standard SAM bridge port.
	SAM_PORT = getEnv("sam_port", "7656")

	// PrimarySessionSwitch enables primary session functionality.
	// This is used internally to enable multi-session capabilities.
	PrimarySessionSwitch = PrimarySessionString()
)

// getEnv retrieves environment variable value with fallback to default.
// This utility function provides a clean way to handle environment variable
// configuration with sensible defaults for SAM bridge connection parameters.
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

// PrimarySessionString returns the primary session configuration identifier.
// This function provides compatibility with the sam3 library's primary session
// switch mechanism for enabling advanced session management features.
func PrimarySessionString() string {
	// Return a basic primary session identifier
	// This will be enhanced when the primary package is implemented
	return "PRIMARY"
}
