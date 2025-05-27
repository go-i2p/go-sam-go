package common

import (
	"net"
	"strings"

	"github.com/samber/oops"
)

// Creates a new controller for the I2P routers SAM bridge.
func OldNewSAM(address string) (*SAM, error) {
	log.WithField("address", address).Debug("Creating new SAM instance")
	var s SAM
	// TODO: clean this up by refactoring the connection setup and error handling logic
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.WithError(err).Error("Failed to dial SAM address")
		return nil, oops.Errorf("error dialing to address '%s': %w", address, err)
	}
	if _, err := conn.Write(s.SAMEmit.HelloBytes()); err != nil {
		log.WithError(err).Error("Failed to write hello message")
		conn.Close()
		return nil, oops.Errorf("error writing to address '%s': %w", address, err)
	}
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read SAM response")
		conn.Close()
		return nil, oops.Errorf("error reading onto buffer: %w", err)
	}
	if strings.Contains(string(buf[:n]), HELLO_REPLY_OK) {
		log.Debug("SAM hello successful")
		s.SAMEmit.I2PConfig.SetSAMAddress(address)
		s.Conn = conn
		resolver, err := NewSAMResolver(&s)
		s.SAMResolver = *resolver
		if err != nil {
			log.WithError(err).Error("Failed to create SAM resolver")
			return nil, oops.Errorf("error creating resolver: %w", err)
		}
		return &s, nil
	} else if string(buf[:n]) == HELLO_REPLY_NOVERSION {
		log.Error("SAM bridge does not support SAMv3")
		conn.Close()
		return nil, oops.Errorf("That SAM bridge does not support SAMv3.")
	} else {
		log.WithField("response", string(buf[:n])).Error("Unexpected SAM response")
		conn.Close()
		return nil, oops.Errorf("%s", string(buf[:n]))
	}
}

func NewSAM(address string) (*SAM, error) {
	logger := log.WithField("address", address)
	logger.Debug("Creating new SAM instance")

	conn, err := connectToSAM(address)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to SAM bridge")
		return nil, err
	}

	s := &SAM{
		Conn: conn,
	}

	if err = sendHelloAndValidate(conn, s); err != nil {
		logger.WithError(err).Error("Failed to send hello and validate SAM connection")
		conn.Close()
		return nil, err
	}

	s.SAMEmit.I2PConfig.SetSAMAddress(address)

	resolver, err := NewSAMResolver(s)
	if err != nil {
		logger.WithError(err).Error("Failed to create SAM resolver")
		return nil, oops.Errorf("failed to create SAM resolver: %w", err)
	}
	s.SAMResolver = *resolver
	logger.Debug("Successfully created new SAM instance")

	return s, nil
}
