package datagram3

import (
	"encoding/base64"
	"net"
	"strings"

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
	buffer := make([]byte, 65536) // Large buffer for UDP datagrams (I2P maximum)
	n, _, err := udpConn.ReadFromUDP(buffer)
	if err != nil {
		return nil, oops.Errorf("failed to read from UDP connection: %w", err)
	}

	log.WithFields(logrus.Fields{
		"bytes_read": n,
		"style":      "DATAGRAM3",
	}).Debug("Received UDP datagram3 message with hash source")

	// Parse the UDP datagram format per SAMv3.md
	response := string(buffer[:n])

	// Find the first newline - that's the end of the header line
	firstNewline := strings.Index(response, "\n")
	if firstNewline == -1 {
		return nil, oops.Errorf("invalid UDP datagram3 format: no newline found")
	}

	// Line 1: Source hash (44-byte base64) followed by optional FROM_PORT=nnn TO_PORT=nnn
	headerLine := strings.TrimSpace(response[:firstNewline])

	if headerLine == "" {
		return nil, oops.Errorf("empty header line in UDP datagram3")
	}

	// Parse the header line to extract the source hash
	// Format: "$hash_base64 FROM_PORT=nnn TO_PORT=nnn"
	// We need to split on space and take the first part as the hash
	parts := strings.Fields(headerLine)
	if len(parts) == 0 {
		return nil, oops.Errorf("empty header line in UDP datagram3")
	}

	hashBase64 := parts[0] // First field is the source hash
	// Remaining parts are FROM_PORT and TO_PORT which we ignore for now

	// Hash MUST be exactly 44 bytes base64 (32 bytes binary)
	if len(hashBase64) != 44 {
		return nil, oops.Errorf("invalid hash length: %d (expected 44 base64 chars)", len(hashBase64))
	}

	// Decode hash from base64 to 32-byte binary
	hashBytes, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		return nil, oops.Errorf("failed to decode source hash from base64: %w", err)
	}

	// Validate decoded hash is exactly 32 bytes
	if len(hashBytes) != 32 {
		return nil, oops.Errorf("invalid hash binary length: %d (expected 32 bytes)", len(hashBytes))
	}

	log.WithFields(logrus.Fields{
		"hash_base64": hashBase64,
		"hash_len":    len(hashBytes),
	}).Debug("Parsed source hash")

	// Everything after the first newline is the payload
	data := response[firstNewline+1:]

	if data == "" {
		return nil, oops.Errorf("no data in UDP datagram3")
	}

	return s.createDatagram(hashBytes, data)
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
