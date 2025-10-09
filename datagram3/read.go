package datagram3

import (
	"net"
	"strings"

	"github.com/go-i2p/common/base64"

	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// readDatagramFromUDP reads a forwarded datagram3 message from the UDP connection.
//
// This is used for receiving datagrams where the SAM bridge forwards messages via UDP.
// DATAGRAM3 uses hash-based source identification instead of full destinations.
//
// Format per SAMv3.md:
//
//	Line 1: $hash (44-byte base64 hash)
//	        [FROM_PORT=nnn] [TO_PORT=nnn] (SAMv3.2+, may be on one or two lines)
//	Then: \n (empty line separator)
//	Remaining: $datagram_payload (raw data)
//
// The hash decodes to 32 bytes binary. To reply:
//  1. Base32-encode hash to 52 characters
//  2. Append ".b32.i2p" suffix
//  3. Use NAMING LOOKUP to get full destination
//  4. Cache result to avoid repeated lookups
func (s *Datagram3Session) readDatagramFromUDP(udpConn *net.UDPConn) (*Datagram3, error) {
	response, n, err := s.readUDPBuffer(udpConn)
	if err != nil {
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"bytes_read": n,
		"style":      "DATAGRAM3",
	}).Debug("Received UDP datagram3 message with hash source")

	headerLine, data, err := s.parseResponseFormat(response)
	if err != nil {
		return nil, err
	}

	hashBytes, err := s.extractAndValidateHash(headerLine)
	if err != nil {
		return nil, err
	}

	return s.createDatagram(hashBytes, data)
}

// readUDPBuffer reads raw data from the UDP connection and returns the response string.
func (s *Datagram3Session) readUDPBuffer(udpConn *net.UDPConn) (string, int, error) {
	buffer := make([]byte, 65536) // Large buffer for UDP datagrams (I2P maximum)
	n, _, err := udpConn.ReadFromUDP(buffer)
	if err != nil {
		return "", 0, oops.Errorf("failed to read from UDP connection: %w", err)
	}
	return string(buffer[:n]), n, nil
}

// parseResponseFormat splits the UDP response into header line and payload data.
func (s *Datagram3Session) parseResponseFormat(response string) (string, string, error) {
	firstNewline := strings.Index(response, "\n")
	if firstNewline == -1 {
		return "", "", oops.Errorf("invalid UDP datagram3 format: no newline found")
	}

	headerLine := strings.TrimSpace(response[:firstNewline])
	if headerLine == "" {
		return "", "", oops.Errorf("empty header line in UDP datagram3")
	}

	data := response[firstNewline+1:]
	if data == "" {
		return "", "", oops.Errorf("no data in UDP datagram3")
	}

	return headerLine, data, nil
}

// extractAndValidateHash parses the header line to extract and validate the source hash.
func (s *Datagram3Session) extractAndValidateHash(headerLine string) ([]byte, error) {
	hashBase64, err := s.parseHashFromHeader(headerLine)
	if err != nil {
		return nil, err
	}

	hashBytes, err := s.decodeAndValidateHash(hashBase64)
	if err != nil {
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"hash_base64": hashBase64,
		"hash_len":    len(hashBytes),
	}).Debug("Parsed source hash")

	return hashBytes, nil
}

// parseHashFromHeader extracts the base64-encoded hash from the header line.
func (s *Datagram3Session) parseHashFromHeader(headerLine string) (string, error) {
	parts := strings.Fields(headerLine)
	if len(parts) == 0 {
		return "", oops.Errorf("empty header line in UDP datagram3")
	}

	hashBase64 := parts[0] // First field is the source hash
	if len(hashBase64) != 44 {
		return "", oops.Errorf("invalid hash length: %d (expected 44 base64 chars)", len(hashBase64))
	}

	return hashBase64, nil
}

// decodeAndValidateHash decodes the base64 hash and validates its binary length.
func (s *Datagram3Session) decodeAndValidateHash(hashBase64 string) ([]byte, error) {
	hashBytes, err := base64.I2PEncoding.DecodeString(hashBase64)
	if err != nil {
		return nil, oops.Errorf("failed to decode source hash from base64: %w", err)
	}

	if len(hashBytes) != 32 {
		return nil, oops.Errorf("invalid hash binary length: %d (expected 32 bytes)", len(hashBytes))
	}

	return hashBytes, nil
}

// createDatagram constructs the final Datagram3 from parsed hash and data.
//
// The datagram is created with:
//   - Data: Raw payload bytes
//   - SourceHash: 32-byte hash
//   - Source: Empty (not resolved, requires NAMING LOOKUP for replies)
//   - Local: This session's I2P address
func (s *Datagram3Session) createDatagram(hashBytes []byte, data string) (*Datagram3, error) {
	// Create datagram with hash
	// Source is empty (not resolved) - applications resolve on-demand for replies
	datagram := &Datagram3{
		Data:       []byte(data),
		SourceHash: hashBytes, // 32-byte hash
		Source:     "",        // Not resolved (empty = requires ResolveSource())
		Local:      s.Addr(),
	}

	log.WithFields(logrus.Fields{
		"data_len":   len(datagram.Data),
		"hash_len":   len(datagram.SourceHash),
		"source_set": datagram.Source != "",
	}).Debug("Created datagram3 with hash source")

	return datagram, nil
}
