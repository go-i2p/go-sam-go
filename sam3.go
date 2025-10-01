// Package sam provides a pure-Go implementation of SAMv3.3 for I2P networks.
// This file implements the main wrapper functions that delegate to sub-package implementations
// while providing the sam3-compatible API surface at the root package level.
package sam

import (
	"strings"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/go-sam-go/primary"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"

	rand "github.com/go-i2p/crypto/rand"
)

// NewSAM creates a new SAM connection to the specified address and performs the initial handshake.
// This is the main entry point for establishing connections to the I2P SAM bridge.
// Address should be in the format "host:port", typically "127.0.0.1:7656" for local I2P routers.
//
// The function connects to the SAM bridge, performs the protocol handshake, and initializes
// the resolver for I2P name lookups. It returns a ready-to-use SAM instance or an error
// if any step of the initialization process fails.
//
// Example:
//
//	sam, err := NewSAM("127.0.0.1:7656")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer sam.Close()
func NewSAM(address string) (*SAM, error) {
	return common.NewSAM(address)
}

// ExtractDest extracts the destination address from a SAM protocol response string.
// This utility function takes the first space-separated token from the input as the destination.
// It's commonly used for parsing SAM session creation responses and connection messages.
//
// Example:
//
//	dest := ExtractDest("ABCD1234...destination_address RESULT=OK")
//	// Returns: "ABCD1234...destination_address"
func ExtractDest(input string) string {
	return common.ExtractDest(input)
}

// ExtractPairInt extracts an integer value from a key=value pair in a space-separated string.
// This utility function searches for the specified key and converts its value to an integer.
// Returns 0 if the key is not found or the value cannot be converted to an integer.
//
// Example:
//
//	port := ExtractPairInt("HOST=example.org PORT=1234 TYPE=stream", "PORT")
//	// Returns: 1234
func ExtractPairInt(input, value string) int {
	return common.ExtractPairInt(input, value)
}

// ExtractPairString extracts a string value from a key=value pair in a space-separated string.
// This utility function searches for the specified key and returns its associated value.
// Returns empty string if the key is not found or has no value.
//
// Example:
//
//	host := ExtractPairString("HOST=example.org PORT=1234 TYPE=stream", "HOST")
//	// Returns: "example.org"
func ExtractPairString(input, value string) string {
	return common.ExtractPairString(input, value)
}

// GenerateOptionString converts a slice of tunnel options into a single space-separated string.
// This utility function takes an array of I2P tunnel configuration options and formats them
// for use in SAM protocol commands. Each option should be in "key=value" format.
//
// Example:
//
//	opts := []string{"inbound.length=3", "outbound.length=3"}
//	result := GenerateOptionString(opts)
//	// Returns: "inbound.length=3 outbound.length=3"
func GenerateOptionString(opts []string) string {
	return strings.Join(opts, " ")
}

// GetSAM3Logger returns the initialized logger instance used by the SAM library.
// This function provides access to the structured logger for applications that want
// to integrate with the library's logging system or adjust log levels.
//
// The logger is configured with appropriate fields for I2P and SAM operations,
// supporting debug, info, warn, and error levels with structured output.
func GetSAM3Logger() *logrus.Logger {
	// Create a new logrus logger that's compatible with the SAM library expectations
	// The go-i2p/logger package uses its own logger type, so we create a logrus instance
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Configure formatter for I2P operations
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return logger
}

// IgnorePortError filters out "missing port in address" errors for convenience when parsing addresses.
// This utility function is used when working with addresses that may not include port numbers.
// Returns nil if the error is about a missing port, otherwise returns the original error unchanged.
//
// This is particularly useful when parsing I2P destination addresses that don't always
// include port specifications, allowing graceful handling of address parsing operations.
//
// Example:
//
//	_, _, err := net.SplitHostPort("example.i2p")  // This would error
//	err = IgnorePortError(err)  // This returns nil
func IgnorePortError(err error) error {
	return common.IgnorePortError(err)
}

// InitializeSAM3Logger configures the logging system for the SAM library.
// This function sets up the logger with appropriate configuration for I2P operations,
// including proper log levels and formatting for SAM protocol debugging.
//
// The logger respects environment variables for configuration:
// - DEBUG_I2P: Controls log level (debug, info, warn, error)
// Applications should call this once during initialization if they want to enable
// structured logging for SAM operations.
func InitializeSAM3Logger() {
	// The go-i2p/logger package handles initialization automatically
	// This function provides compatibility with the sam3 API expectations
	log := GetSAM3Logger()
	log.Info("SAM3 logger initialized")
}

// RandString generates a random string suitable for use as session identifiers or tunnel names.
// This utility function creates cryptographically secure random strings using I2P's
// random number generator. The generated strings are URL-safe and suitable for use
// in SAM protocol commands and session identification.
//
// Returns a random string that can be used for session IDs, tunnel names, or other
// identifiers that require uniqueness and unpredictability in I2P operations.
func RandString() string {
	// Use a simple but secure approach for generating random session identifiers
	// Generate a 12-character random string using lowercase letters (similar to tunnel names)
	const (
		nameLength = 12
		letters    = "abcdefghijklmnopqrstuvwxyz0123456789"
	)

	result := make([]byte, nameLength)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}

	return string(result)
}

// SAMDefaultAddr constructs the default SAM bridge address with fallback support.
// This utility function provides a standardized way to determine the SAM bridge address,
// using the provided fallback if the standard environment variables are not set.
//
// The function checks SAM_HOST and SAM_PORT variables first, then falls back to the
// provided fallforward parameter if those are not available. This enables flexible
// configuration while providing sensible defaults for most I2P installations.
//
// Example:
//
//	addr := SAMDefaultAddr("127.0.0.1:7656")
//	// Returns: "127.0.0.1:7656" (or values from SAM_HOST/SAM_PORT if set)
func SAMDefaultAddr(fallforward string) string {
	// Use the global variables that are already configured with environment support
	if SAM_HOST != "" && SAM_PORT != "" {
		return SAM_HOST + ":" + SAM_PORT
	}
	return fallforward
}

// SplitHostPort separates host and port from a combined address string with I2P-aware handling.
// Unlike net.SplitHostPort, this function handles I2P addresses gracefully, including those
// without explicit port specifications. Returns host, port as strings, and error.
//
// This function is I2P-aware and handles the common case where I2P destination addresses
// don't include port numbers. Port defaults to "0" if not specified, and the function
// uses IgnorePortError internally to handle missing port situations gracefully.
//
// Example:
//
//	host, port, err := SplitHostPort("example.i2p")
//	// Returns: "example.i2p", "0", nil
func SplitHostPort(hostport string) (string, string, error) {
	return common.SplitHostPort(hostport)
}

// NewSAMResolver creates a new SAM resolver instance for I2P name lookups.
// This function creates a resolver that can translate I2P names (like "example.i2p")
// into Base32 destination addresses for use in connections and messaging.
//
// The resolver uses the provided SAM connection for performing lookups through the
// I2P network's address book and naming services. It's essential for applications
// that want to connect to I2P services using human-readable names.
//
// Example:
//
//	resolver, err := NewSAMResolver(sam)
//	if err != nil {
//	    return err
//	}
//	addr, err := resolver.Resolve("example.i2p")
func NewSAMResolver(parent *SAM) (*SAMResolver, error) {
	return common.NewSAMResolver(parent)
}

// NewFullSAMResolver creates a new complete SAM resolver by establishing its own connection.
// This convenience function creates both a SAM connection and resolver in a single operation.
// It's useful when you only need name resolution and don't require a persistent SAM connection
// for session management or other operations.
//
// The resolver will establish its own connection to the specified address and be ready
// for immediate use. The caller is responsible for closing the resolver when done.
//
// Example:
//
//	resolver, err := NewFullSAMResolver("127.0.0.1:7656")
//	if err != nil {
//	    return err
//	}
//	defer resolver.Close()
func NewFullSAMResolver(address string) (*SAMResolver, error) {
	sam, err := NewSAM(address)
	if err != nil {
		return nil, oops.Errorf("failed to create SAM connection for resolver: %w", err)
	}

	resolver, err := common.NewSAMResolver(sam)
	if err != nil {
		sam.Close()
		return nil, oops.Errorf("failed to create SAM resolver: %w", err)
	}

	return resolver, nil
}

// NewPrimarySession creates a new primary session with the SAM bridge using default settings.
// This method establishes a new primary session for managing multiple sub-sessions over I2P
// with the specified session ID, cryptographic keys, and configuration options. It uses default
// signature settings and provides a simple interface for basic primary session needs.
//
// The primary session acts as a master container that can create and manage multiple sub-sessions
// of different types (stream, datagram, raw) while sharing the same I2P identity and tunnel
// infrastructure for enhanced efficiency and consistent anonymity properties.
//
// Example:
//
//	sam, err := NewSAM("127.0.0.1:7656")
//	if err != nil {
//	    return err
//	}
//	defer sam.Close()
//
//	keys, err := sam.NewKeys()
//	if err != nil {
//	    return err
//	}
//
//	session, err := NewPrimarySession(sam, "my-primary", keys, []string{"inbound.length=2"})
//	if err != nil {
//	    return err
//	}
//	defer session.Close()
//
//	// Create sub-sessions
//	streamSub, err := session.NewStreamSubSession("stream-1", streamOptions)
//	datagramSub, err := session.NewDatagramSubSession("datagram-1", datagramOptions)
func NewPrimarySession(sam *SAM, id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error) {
	return primary.NewPrimarySession(sam, id, keys, options)
}

// NewPrimarySessionWithSignature creates a new primary session with custom signature type.
// This method allows specifying a custom cryptographic signature type for the session,
// enabling advanced security configurations beyond the default signature algorithm.
// Different signature types provide various security levels, compatibility options,
// and performance characteristics for different I2P network requirements.
//
// The primary session created with custom signature maintains the same multi-session
// management capabilities while using the specified cryptographic parameters for
// enhanced security or compatibility with specific I2P network configurations.
//
// Example:
//
//	session, err := NewPrimarySessionWithSignature(sam, id, keys, options, "EdDSA_SHA512_Ed25519")
//	if err != nil {
//	    return err
//	}
//	defer session.Close()
//
//	datagramSub, err := session.NewDatagramSubSession("datagram-1", datagramOptions)
func NewPrimarySessionWithSignature(sam *SAM, id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error) {
	primarySAM := &primary.SAM{SAM: sam}
	return primarySAM.NewPrimarySessionWithSignature(id, keys, options, sigType)
}

// NewPrimarySessionWithPorts creates a new primary session with port specifications.
// This method allows configuring specific port ranges for the session, enabling fine-grained
// control over network communication ports for advanced routing scenarios. Port configuration
// is useful for applications requiring specific port mappings, firewall compatibility,
// or integration with existing network infrastructure and service discovery mechanisms.
//
// The primary session created with port configuration maintains full multi-session management
// capabilities while using the specified port parameters for network communication optimization
// and compatibility with existing network configurations or security requirements.
//
// Example:
//
//	session, err := NewPrimarySessionWithPorts(sam, id, "8080", "8081", keys, options)
//	if err != nil {
//	    return err
//	}
//	defer session.Close()
//
//	rawSub, err := session.NewRawSubSession("raw-1", rawOptions)
func NewPrimarySessionWithPorts(sam *SAM, id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error) {
	primarySAM := &primary.SAM{SAM: sam}
	return primarySAM.NewPrimarySessionWithPorts(id, fromPort, toPort, keys, options)
}
