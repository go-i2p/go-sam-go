package common

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/samber/oops"
	"github.com/sirupsen/logrus"

	rand "github.com/go-i2p/crypto/rand"
)

// IgnorePortError filters out "missing port in address" errors for convenience.
// This function is used when parsing addresses that may not include port numbers.
// Returns nil if the error is about missing port, otherwise returns the original error.
func IgnorePortError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "missing port in address") {
		log.Debug("Ignoring 'missing port in address' error")
		err = nil
	}
	return err
}

// SplitHostPort separates host and port from a combined address string.
// Unlike net.SplitHostPort, this function handles addresses without ports gracefully.
// Returns host, port as strings, and error. Port defaults to "0" if not specified.
func SplitHostPort(hostport string) (string, string, error) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		if IgnorePortError(err) == nil {
			log.WithField("host", hostport).Debug("Using full string as host, port set to 0")
			host = hostport
			port = "0"
		}
	}
	log.WithFields(logrus.Fields{
		"host": host,
		"port": port,
	}).Debug("Split host and port")
	return host, port, nil
}

// ExtractPairString extracts the value from a key=value pair in a space-separated string.
// Searches for the specified key prefix and returns the associated value.
// Returns empty string if the key is not found or has no value.
func ExtractPairString(input, value string) string {
	log.WithFields(logrus.Fields{"input": input, "value": value}).Debug("ExtractPairString called")
	parts := strings.Split(input, " ")
	for _, part := range parts {
		log.WithField("part", part).Debug("Checking part")
		if strings.HasPrefix(part, value) {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				log.WithFields(logrus.Fields{"key": kv[0], "value": kv[1]}).Debug("Pair extracted")
				return kv[1]
			}
		}
	}
	log.WithFields(logrus.Fields{"input": input, "value": value}).Debug("No pair found")
	return ""
}

// ExtractPairInt extracts an integer value from a key=value pair in a space-separated string.
// Uses ExtractPairString internally and converts the result to integer.
// Returns 0 if the key is not found or the value cannot be converted to integer.
func ExtractPairInt(input, value string) int {
	rv, err := strconv.Atoi(ExtractPairString(input, value))
	if err != nil {
		log.WithFields(logrus.Fields{"input": input, "value": value}).Debug("No pair found")
		return 0
	}
	log.WithField("result", rv).Debug("Pair extracted and converted to int")
	return rv
}

// ExtractDest extracts the destination address from a SAM protocol response.
// Takes the first space-separated token from the input string as the destination.
// Used for parsing SAM session creation responses and connection messages.
func ExtractDest(input string) string {
	log.WithField("input", input).Debug("ExtractDest called")
	dest := strings.Split(input, " ")[0]
	log.WithField("dest", dest).Debug("Destination extracted")
	return dest
}

// RandPort generates a random available port number for local testing.
// Attempts to find a port that is available for both TCP and UDP connections.
// Returns the port as a string or an error if no available port is found after 30 attempts.
func RandPort() (portNumber string, err error) {
	maxAttempts := 30
	for range maxAttempts {
		port, err := generateRandomPort()
		if err != nil {
			return "", err
		}

		if isPortAvailable(port) {
			return port, nil
		}
	}

	return "", oops.Errorf("unable to find a pair of available tcp and udp ports in %v attempts", maxAttempts)
}

// generateRandomPort creates a random port number in the range 10000-65534.
// Uses crypto/rand for thread-safe random generation.
func generateRandomPort() (string, error) {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", oops.Wrapf(err, "failed to generate random bytes")
	}

	// Convert to uint32 and scale to port range (10000-65534)
	p := int(binary.BigEndian.Uint32(buf[:]))%55534 + 10000
	return strconv.Itoa(p), nil
}

// isPortAvailable checks if a port is available for both TCP and UDP connections.
// Returns true if the port can be bound on both protocols, false otherwise.
func isPortAvailable(port string) bool {
	return isTCPPortAvailable(port) && isUDPPortAvailable(port)
}

// isTCPPortAvailable checks if a TCP port is available for binding.
// Returns true if the port can be bound, false otherwise.
func isTCPPortAvailable(port string) bool {
	l, err := net.Listen("tcp", net.JoinHostPort("localhost", port))
	if err != nil {
		return false
	}
	defer l.Close()
	return true
}

// isUDPPortAvailable checks if a UDP port is available for binding.
// Returns true if the port can be bound, false otherwise.
func isUDPPortAvailable(port string) bool {
	l, err := net.Listen("udp", net.JoinHostPort("localhost", port))
	if err != nil {
		return false
	}
	defer l.Close()
	return true
}

// generateRandomTunnelName creates a random 12-character tunnel name using lowercase letters.
// This function is deterministic for testing when a fixed seed is used.
func (f *I2PConfig) generateRandomTunnelName() string {
	const (
		nameLength = 12
		letters    = "abcdefghijklmnopqrstuvwxyz"
	)

	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	name := make([]byte, nameLength)

	for i := range name {
		name[i] = letters[generator.Intn(len(letters))]
	}

	return string(name)
}

// validateEncryptionTypes checks that all comma-separated values are valid integers
func (f *I2PConfig) validateEncryptionTypes(encTypes string) error {
	for _, s := range strings.Split(encTypes, ",") {
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			return fmt.Errorf("empty encryption type")
		}
		if _, err := strconv.Atoi(trimmed); err != nil {
			return fmt.Errorf("invalid encryption type '%s': %w", trimmed, err)
		}
	}
	return nil
}

// formatLeaseSetEncryptionType creates the formatted configuration string
func (f *I2PConfig) formatLeaseSetEncryptionType(encType string) string {
	log.WithField("leaseSetEncType", encType).Debug("Lease set encryption type set")
	return fmt.Sprintf("i2cp.leaseSetEncType=%s", encType)
}

// collectTunnelSettings returns all tunnel-related configuration strings
func (f *I2PConfig) collectTunnelSettings() []string {
	return []string{
		f.InboundLength(),
		f.OutboundLength(),
		f.InboundLengthVariance(),
		f.OutboundLengthVariance(),
		f.InboundBackupQuantity(),
		f.OutboundBackupQuantity(),
		f.InboundQuantity(),
		f.OutboundQuantity(),
	}
}

// collectConnectionSettings returns all connection behavior configuration strings
func (f *I2PConfig) collectConnectionSettings() []string {
	return []string{
		f.UsingCompression(),
		f.DoZero(),      // Zero hop settings
		f.Reduce(),      // Reduce idle settings
		f.Close(),       // Close idle settings
		f.Reliability(), // Message reliability
	}
}

// collectLeaseSetSettings returns all lease set configuration strings
func (f *I2PConfig) collectLeaseSetSettings() []string {
	lsk, lspk, lspsk := f.LeaseSetSettings()
	return []string{
		f.EncryptLease(), // Lease encryption
		lsk, lspk, lspsk, // Lease set keys
		f.LeaseSetEncryptionType(), // Lease set encryption type
	}
}

// collectAccessSettings returns all access control configuration strings
func (f *I2PConfig) collectAccessSettings() []string {
	return []string{
		f.Accesslisttype(), // Access list type
		f.Accesslist(),     // Access list
	}
}
