package common

import (
	"net"
	"strings"

	"github.com/samber/oops"
)

// connectToSAM establishes a TCP connection to the SAM bridge at the specified address.
// This is an internal helper function used during SAM instance initialization.
// Returns the established connection or an error if the connection fails.
func connectToSAM(address string) (net.Conn, error) {
	log.WithField("address", address).Debug("Connecting to SAM bridge")

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.WithField("address", address).WithError(err).Error("Failed to connect to SAM bridge")
		return nil, oops.Errorf("failed to connect to SAM bridge at %s: %w", address, err)
	}

	log.WithField("address", address).Debug("Successfully connected to SAM bridge")
	return conn, nil
}

// sendHelloAndValidate performs the SAM protocol handshake and validates the response.
// This internal function sends the HELLO message and ensures the SAM bridge supports the protocol.
// Returns an error if the handshake fails or the protocol version is unsupported.
func sendHelloAndValidate(conn net.Conn, s *SAM) error {
	helloMessage := s.SAMEmit.HelloBytes()
	log.WithField("message", string(helloMessage)).Debug("Sending HELLO message to SAM bridge")

	if _, err := conn.Write(helloMessage); err != nil {
		log.WithError(err).Error("Failed to send HELLO message")
		return oops.Errorf("failed to send hello message: %w", err)
	}

	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read SAM HELLO response")
		return oops.Errorf("failed to read SAM response: %w", err)
	}

	response := string(buf[:n])
	log.WithField("response", response).Debug("Received SAM HELLO response")

	switch {
	case strings.Contains(response, HELLO_REPLY_OK):
		log.Debug("SAM hello successful")
		return nil
	case response == HELLO_REPLY_NOVERSION:
		log.Error("SAM bridge does not support SAMv3")
		return oops.Errorf("SAM bridge does not support SAMv3")
	default:
		log.WithField("response", response).Error("Unexpected SAM response")
		return oops.Errorf("unexpected SAM response: %s", response)
	}
}
