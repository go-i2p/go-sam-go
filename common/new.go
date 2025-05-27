package common

import (
	"net"
	"strings"

	"github.com/samber/oops"
)

// NewSAM creates a new SAM instance by connecting to the specified address,
// performing the hello handshake, and initializing the SAM resolver.
// It returns a pointer to the SAM instance or an error if any step fails.
// This function combines connection establishment and hello handshake into a single step,
// eliminating the need for separate helper functions.
// It also initializes the SAM resolver directly after the connection is established.
// The SAM instance is ready to use for further operations like session creation or name resolution.
func NewSAM(address string) (*SAM, error) {
	logger := log.WithField("address", address)
	logger.Debug("Creating new SAM instance")

	// Inline connection establishment - eliminates connectToSAM helper
	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to SAM bridge")
		return nil, oops.Errorf("failed to connect to SAM bridge at %s: %w", address, err)
	}

	s := &SAM{
		Conn: conn,
	}

	// Inline hello handshake - eliminates sendHelloAndValidate helper
	if _, err := conn.Write(s.SAMEmit.HelloBytes()); err != nil {
		logger.WithError(err).Error("Failed to send hello message")
		conn.Close()
		return nil, oops.Errorf("failed to send hello message: %w", err)
	}

	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		logger.WithError(err).Error("Failed to read SAM response")
		conn.Close()
		return nil, oops.Errorf("failed to read SAM response: %w", err)
	}

	response := string(buf[:n])
	switch {
	case strings.Contains(response, HELLO_REPLY_OK):
		logger.Debug("SAM hello successful")
	case response == HELLO_REPLY_NOVERSION:
		logger.Error("SAM bridge does not support SAMv3")
		conn.Close()
		return nil, oops.Errorf("SAM bridge does not support SAMv3")
	default:
		logger.WithField("response", response).Error("Unexpected SAM response")
		conn.Close()
		return nil, oops.Errorf("unexpected SAM response: %s", response)
	}

	s.SAMEmit.I2PConfig.SetSAMAddress(address)

	resolver, err := NewSAMResolver(s)
	if err != nil {
		logger.WithError(err).Error("Failed to create SAM resolver")
		conn.Close()
		return nil, oops.Errorf("failed to create SAM resolver: %w", err)
	}
	s.SAMResolver = *resolver
	logger.Debug("Successfully created new SAM instance")

	return s, nil
}
