package datagram3

import (
	"fmt"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// SetTimeout sets the timeout for datagram3 write operations.
// This method configures the maximum time to wait for datagram send operations
// to complete. The timeout prevents indefinite blocking during network congestion or connection
// issues. Returns the writer instance for method chaining convenience.
//
// Example usage:
//
//	writer.SetTimeout(30*time.Second).SendDatagram(data, destination)
func (w *Datagram3Writer) SetTimeout(timeout time.Duration) *Datagram3Writer {
	// Configure the timeout for send operations to prevent indefinite blocking
	w.timeout = timeout
	return w
}

// SendDatagram sends a datagram to the specified I2P destination.
//
// This method uses the SAMv3 UDP approach: sending via UDP socket to port 7655 with DATAGRAM3 format.
// The destination can be:
//   - Full base64 destination (516+ chars)
//   - Hostname (.i2p address)
//   - B32 address (52 chars + .b32.i2p)
//   - B32 address derived from received DATAGRAM3 hash (via ResolveSource)
//
// Maximum datagram size is 31744 bytes total (including headers), with 11 KB recommended for
// best reliability across the I2P network. It blocks until the datagram is sent or an error occurs,
// respecting the configured timeout.
//
// Example usage:
//
//	// Send to full destination
//	err := writer.SendDatagram([]byte("hello world"), destinationAddr)
//
//	// Reply to received datagram (requires hash resolution)
//	if err := receivedDatagram.ResolveSource(session); err != nil {
//	    return err
//	}
//	err := writer.SendDatagram([]byte("reply"), receivedDatagram.Source)
func (w *Datagram3Writer) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error {
	// Check if the session is closed before attempting to send
	w.session.mu.RLock()
	if w.session.closed {
		w.session.mu.RUnlock()
		return oops.Errorf("session is closed")
	}
	w.session.mu.RUnlock()

	// Create detailed logging context for debugging send operations
	logger := log.WithFields(logrus.Fields{
		"session_id":  w.session.ID(),
		"destination": dest.Base32(),
		"size":        len(data),
		"style":       "DATAGRAM3",
	})
	logger.Debug("Sending datagram3 message via UDP socket")

	// Use UDP socket approach (SAMv3 method) to send DATAGRAM3 messages
	// Connect to SAM's UDP port (default 7655) for datagram3 transmission
	samHost := w.session.sam.SAMEmit.I2PConfig.SamHost
	if samHost == "" {
		samHost = "127.0.0.1" // Default SAM host
	}
	samUDPPort := "7655" // Default SAM UDP port for datagram3 transmission

	udpAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(samHost, samUDPPort))
	if err != nil {
		logger.WithError(err).Error("Failed to resolve SAM UDP address")
		return oops.Errorf("failed to resolve SAM UDP address: %w", err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to SAM UDP port")
		return oops.Errorf("failed to connect to SAM UDP port: %w", err)
	}
	defer udpConn.Close()

	// Construct the SAMv3 UDP datagram3 format:
	// First line: "3.3 <session_id> <destination> [options]\n"
	// Remaining data: the actual message payload
	sessionID := w.session.ID()
	destination := dest.Base64()

	// Create the header line according to SAMv3 specification
	// DATAGRAM3 uses same format as DATAGRAM/DATAGRAM2 for sending
	// Only reception format differs (hash-based source)
	headerLine := fmt.Sprintf("3.3 %s %s\n", sessionID, destination)

	// Combine header and data into final UDP packet
	udpMessage := append([]byte(headerLine), data...)

	logger.WithFields(logrus.Fields{
		"header":     headerLine,
		"total_size": len(udpMessage),
	}).Debug("Sending UDP datagram3 to SAM")

	// Send the datagram3 message via UDP to SAM bridge
	_, err = udpConn.Write(udpMessage)
	if err != nil {
		logger.WithError(err).Error("Failed to send UDP datagram3 to SAM")
		return oops.Errorf("failed to send UDP datagram3 to SAM: %w", err)
	}

	logger.Debug("Successfully sent datagram3 message via UDP")
	return nil
}

// ReplyToDatagram sends a reply to a received DATAGRAM3 message.
//
// This automatically resolves the source hash if not already resolved, then sends the reply.
// The source hash is resolved via NAMING LOOKUP and cached to avoid repeated lookups.
//
// Example usage:
//
//	// Receive datagram
//	dg, err := reader.ReceiveDatagram()
//	if err != nil {
//	    return err
//	}
//
//	// Reply (automatically resolves hash)
//	writer := session.NewWriter()
//	err = writer.ReplyToDatagram([]byte("reply"), dg)
func (w *Datagram3Writer) ReplyToDatagram(data []byte, original *Datagram3) error {
	// Ensure source is resolved (performs NAMING LOOKUP if needed)
	if original.Source == "" {
		log.Debug("Source not resolved, performing hash resolution for reply")
		if err := original.ResolveSource(w.session); err != nil {
			return oops.Errorf("failed to resolve source hash for reply: %w", err)
		}
	}

	// Send to resolved source
	return w.SendDatagram(data, original.Source)
}
