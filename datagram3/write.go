package datagram3

import (
	"fmt"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/go-i2p/logger"
	"github.com/samber/oops"
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
	// Validate session state before attempting send
	if err := w.validateSessionState(); err != nil {
		return err
	}

	// Create logging context for debugging
	log := w.createSendLogger(dest, len(data))
	log.Debug("Sending datagram3 message via UDP socket")

	// Establish UDP connection to SAM bridge
	udpConn, err := w.connectToSAMUDP(log)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	// Build and send the datagram3 message
	udpMessage := w.buildDatagram3Message(dest, data)

	log.WithFields(logger.Fields{
		"total_size": len(udpMessage),
	}).Debug("Sending UDP datagram3 to SAM")

	// Transmit the datagram3 message via UDP to SAM bridge
	if err := w.writeDatagram(udpConn, udpMessage, log); err != nil {
		return err
	}

	log.Debug("Successfully sent datagram3 message via UDP")
	return nil
}

// validateSessionState checks if the session is closed before attempting operations.
// This method verifies the session state with appropriate mutex locking to prevent
// operations on closed sessions. Returns an error if the session is closed.
func (w *Datagram3Writer) validateSessionState() error {
	w.session.mu.RLock()
	defer w.session.mu.RUnlock()

	if w.session.closed {
		return oops.Errorf("session is closed")
	}
	return nil
}

// createSendLogger creates a structured logger with datagram3 send operation context.
// This method configures logging fields including session ID, destination, message size,
// and protocol style for comprehensive debugging of send operations.
func (w *Datagram3Writer) createSendLogger(dest i2pkeys.I2PAddr, dataSize int) *logger.Entry {
	return log.WithFields(logger.Fields{
		"session_id":  w.session.ID(),
		"destination": dest.Base32(),
		"size":        dataSize,
		"style":       "DATAGRAM3",
	})
}

// connectToSAMUDP establishes a UDP connection to the SAM bridge for datagram3 transmission.
// This method resolves the SAM host and UDP port (default 7655), creates the UDP address,
// and establishes the connection. Returns the UDP connection and any error encountered.
func (w *Datagram3Writer) connectToSAMUDP(logger *logger.Entry) (*net.UDPConn, error) {
	// Determine SAM host from configuration or use default
	samHost := w.session.sam.SAMEmit.I2PConfig.SamHost
	if samHost == "" {
		samHost = "127.0.0.1"
	}
	samUDPPort := "7655" // Default SAM UDP port for datagram3

	// Resolve the UDP address for SAM bridge
	udpAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(samHost, samUDPPort))
	if err != nil {
		logger.WithError(err).Error("Failed to resolve SAM UDP address")
		return nil, oops.Errorf("failed to resolve SAM UDP address: %w", err)
	}

	// Establish UDP connection to SAM bridge
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to SAM UDP port")
		return nil, oops.Errorf("failed to connect to SAM UDP port: %w", err)
	}

	return udpConn, nil
}

// buildDatagram3Message constructs the SAMv3 UDP datagram3 message format.
// This method creates the protocol header and combines it with the data payload.
// The header format is: "3.3 <session_id> <destination>\n" followed by the message data.
// Returns the complete UDP message ready for transmission.
func (w *Datagram3Writer) buildDatagram3Message(dest i2pkeys.I2PAddr, data []byte) []byte {
	sessionID := w.session.ID()
	destination := dest.Base64()

	// Create SAMv3 DATAGRAM3 header line
	headerLine := fmt.Sprintf("3.3 %s %s\n", sessionID, destination)

	// Combine header and data into final UDP packet
	return append([]byte(headerLine), data...)
}

// writeDatagram transmits the datagram3 message via UDP to the SAM bridge.
// This method writes the complete UDP message to the connection and handles
// any transmission errors with appropriate logging.
func (w *Datagram3Writer) writeDatagram(udpConn *net.UDPConn, udpMessage []byte, logger *logger.Entry) error {
	_, err := udpConn.Write(udpMessage)
	if err != nil {
		logger.WithError(err).Error("Failed to send UDP datagram3 to SAM")
		return oops.Errorf("failed to send UDP datagram3 to SAM: %w", err)
	}
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
