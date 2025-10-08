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
// ⚠️  CRITICAL SECURITY WARNING: Source is a 44-byte base64 hash, NOT authenticated!
// ⚠️  The hash can be spoofed by malicious actors - do NOT trust without verification.
// ⚠️  This is fundamentally different from DATAGRAM/DATAGRAM2 authenticated sources.
//
// This is used for receiving datagrams where the SAM bridge forwards messages via UDP.
// DATAGRAM3 uses hash-based source identification instead of full destinations.
//
// Format per SAMv3.md:
//
//	Line 1: $hash (44-byte base64 hash, UNAUTHENTICATED!)
//	        [FROM_PORT=nnn] [TO_PORT=nnn] (SAMv3.2+, may be on one or two lines)
//	Then: \n (empty line separator)
//	Remaining: $datagram_payload (raw data)
//
// CRITICAL DIFFERENCE from DATAGRAM/DATAGRAM2:
//   - Source is 44-byte base64 hash (32 bytes binary)
//   - NOT a full destination (516+ chars)
//   - Hash is UNAUTHENTICATED and spoofable
//   - Must use NAMING LOOKUP to resolve hash for replies
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
	}).Debug("Received UDP datagram3 message with UNAUTHENTICATED hash source")

	// Parse the UDP datagram format per SAMv3.md
	response := string(buffer[:n])

	// Find the first newline - that's the end of the header line
	firstNewline := strings.Index(response, "\n")
	if firstNewline == -1 {
		return nil, oops.Errorf("invalid UDP datagram3 format: no newline found")
	}

	// Line 1: Source hash (44-byte base64, UNAUTHENTICATED!) followed by optional FROM_PORT=nnn TO_PORT=nnn
	headerLine := strings.TrimSpace(response[:firstNewline])

	if headerLine == "" {
		return nil, oops.Errorf("empty header line in UDP datagram3")
	}

	// Parse the header line to extract the UNAUTHENTICATED source hash
	// Format: "$hash_base64 FROM_PORT=nnn TO_PORT=nnn"
	// We need to split on space and take the first part as the hash
	parts := strings.Fields(headerLine)
	if len(parts) == 0 {
		return nil, oops.Errorf("empty header line in UDP datagram3")
	}

	hashBase64 := parts[0] // First field is the UNAUTHENTICATED source hash
	// Remaining parts are FROM_PORT and TO_PORT which we ignore for now

	// CRITICAL VALIDATION: Hash MUST be exactly 44 bytes base64 (32 bytes binary)
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
	}).Debug("Parsed UNAUTHENTICATED source hash")

	// Everything after the first newline is the payload
	data := response[firstNewline+1:]

	if data == "" {
		return nil, oops.Errorf("no data in UDP datagram3")
	}

	return s.createDatagram(hashBytes, data)
}

// createDatagram constructs the final Datagram3 from parsed UNAUTHENTICATED hash and data.
//
// ⚠️  SECURITY WARNING: The hash is NOT authenticated and can be spoofed!
// ⚠️  Do NOT trust the source without additional verification.
//
// The datagram is created with:
//   - Data: Raw payload bytes
//   - SourceHash: 32-byte UNAUTHENTICATED hash (spoofable!)
//   - Source: Empty (not resolved, requires NAMING LOOKUP for replies)
//   - Local: This session's I2P address
//
// Applications must call ResolveSource() to convert hash to full destination for replies.
func (s *Datagram3Session) createDatagram(hashBytes []byte, data string) (*Datagram3, error) {
	// Create datagram with UNAUTHENTICATED hash
	// Source is empty (not resolved) - applications resolve on-demand for replies
	datagram := &Datagram3{
		Data:       []byte(data),
		SourceHash: hashBytes, // 32-byte UNAUTHENTICATED hash (spoofable!)
		Source:     "",        // Not resolved (empty = requires ResolveSource())
		Local:      s.Addr(),
	}

	log.WithFields(logrus.Fields{
		"data_len":   len(datagram.Data),
		"hash_len":   len(datagram.SourceHash),
		"source_set": datagram.Source != "",
	}).Debug("Created datagram3 with UNAUTHENTICATED hash source")

	return datagram, nil
}
