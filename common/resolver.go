package common

import (
	"bufio"
	"bytes"
	"errors"
	"strings"

	"github.com/go-i2p/i2pkeys"
)

// NewSAMResolver creates a new SAMResolver using an existing SAM instance.
// This allows sharing a single SAM connection for both session management and address resolution.
// Returns a configured resolver ready for performing I2P address lookups.
func NewSAMResolver(parent *SAM) (*SAMResolver, error) {
	log.Debug("Creating new SAMResolver from existing SAM instance")
	var s SAMResolver
	s.SAM = parent
	return &s, nil
}

// NewFullSAMResolver creates a complete SAMResolver with its own SAM connection.
// Establishes a new connection to the specified SAM bridge address for address resolution.
// Returns a fully configured resolver or an error if connection fails.
func NewFullSAMResolver(address string) (*SAMResolver, error) {
	log.WithField("address", address).Debug("Creating new full SAMResolver")
	var s SAMResolver
	// var err error
	sam, err := NewSAM(address)
	s.SAM = sam
	if err != nil {
		log.WithError(err).Error("Failed to create new SAM instance")
		return nil, err
	}
	return &s, nil
}

// Performs a lookup, probably this order: 1) routers known addresses, cached
// addresses, 3) by asking peers in the I2P network.
func (sam *SAMResolver) Resolve(name string) (i2pkeys.I2PAddr, error) {
	log.WithField("name", name).Debug("Resolving name")

	if err := sam.sendLookupRequest(name); err != nil {
		return i2pkeys.I2PAddr(""), err
	}

	response, err := sam.readLookupResponse()
	if err != nil {
		return i2pkeys.I2PAddr(""), err
	}

	scanner, err := sam.prepareLookupScanner(response)
	if err != nil {
		return i2pkeys.I2PAddr(""), err
	}

	return sam.processLookupResponse(scanner, name)
}

// sendLookupRequest sends a NAMING LOOKUP request to the SAM connection.
// It writes the lookup command and handles any connection errors.
func (sam *SAMResolver) sendLookupRequest(name string) error {
	if _, err := sam.Conn.Write([]byte("NAMING LOOKUP NAME=" + name + "\r\n")); err != nil {
		log.WithError(err).Error("Failed to write to SAM connection")
		sam.Close()
		return err
	}
	return nil
}

// readLookupResponse reads the response from the SAM connection.
// It handles reading errors and connection cleanup on failure.
func (sam *SAMResolver) readLookupResponse() ([]byte, error) {
	buf := make([]byte, 4096)
	n, err := sam.Conn.Read(buf)
	if err != nil {
		log.WithError(err).Error("Failed to read from SAM connection")
		sam.Close()
		return nil, err
	}
	return buf[:n], nil
}

// prepareLookupScanner validates the response format and creates a scanner.
// It ensures the response has the correct "NAMING REPLY" prefix and length.
func (sam *SAMResolver) prepareLookupScanner(response []byte) (*bufio.Scanner, error) {
	if len(response) <= 13 || !strings.HasPrefix(string(response), "NAMING REPLY ") {
		log.Error("Failed to parse SAM response")
		return nil, errors.New("failed to parse SAM response")
	}

	scanner := bufio.NewScanner(bytes.NewReader(response[13:]))
	scanner.Split(bufio.ScanWords)
	return scanner, nil
}

// processLookupResponse processes the scanner tokens and returns the resolved address.
// It handles different response types and accumulates error messages.
func (sam *SAMResolver) processLookupResponse(scanner *bufio.Scanner, name string) (i2pkeys.I2PAddr, error) {
	errStr := ""
	for scanner.Scan() {
		text := scanner.Text()
		log.WithField("text", text).Debug("Parsing SAM response token")

		if resolved, found := sam.handleValueResponse(text); found {
			return resolved, nil
		}

		if sam.shouldSkipToken(text, name) {
			continue
		}

		errStr = sam.handleErrorResponse(text, name, errStr)
	}
	return i2pkeys.I2PAddr(""), errors.New(errStr)
}

// handleValueResponse processes VALUE= responses and returns the resolved address.
// Returns the address and true if this was a VALUE response, empty address and false otherwise.
func (sam *SAMResolver) handleValueResponse(text string) (i2pkeys.I2PAddr, bool) {
	if strings.HasPrefix(text, "VALUE=") {
		addr := i2pkeys.I2PAddr(text[6:])
		log.WithField("addr", addr).Debug("Name resolved successfully")
		return addr, true
	}
	return i2pkeys.I2PAddr(""), false
}

// shouldSkipToken determines if a token should be skipped without processing.
// Returns true for OK results and NAME= tokens that match the requested name.
func (sam *SAMResolver) shouldSkipToken(text, name string) bool {
	return text == SAM_RESULT_OK || text == "NAME="+name
}

// handleErrorResponse processes error responses and accumulates error messages.
// Returns the updated error string with any new error information.
func (sam *SAMResolver) handleErrorResponse(text, name, errStr string) string {
	if text == SAM_RESULT_INVALID_KEY {
		errStr += "Invalid key - resolver."
		log.Error("Invalid key in resolver")
	} else if text == SAM_RESULT_KEY_NOT_FOUND {
		errStr += "Unable to resolve " + name
		log.WithField("name", name).Error("Unable to resolve name")
	} else if strings.HasPrefix(text, "MESSAGE=") {
		errStr += " " + text[8:]
		log.WithField("message", text[8:]).Warn("Received message from SAM")
	}
	return errStr
}
