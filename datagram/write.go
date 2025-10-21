package datagram

import (
	"fmt"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/go-i2p/logger"
	"github.com/samber/oops"
)

// SetTimeout sets the timeout for datagram write operations.
// This method configures the maximum time to wait for datagram send operations to complete.
// The timeout prevents indefinite blocking during network congestion or connection issues.
// Returns the writer instance for method chaining convenience.
// Example usage: writer.SetTimeout(30*time.Second).SendDatagram(data, destination)
func (w *DatagramWriter) SetTimeout(timeout time.Duration) *DatagramWriter {
	// Configure the timeout for send operations to prevent indefinite blocking
	w.timeout = timeout
	return w
}

// SendDatagram sends a datagram to the specified I2P destination address.
// This method uses the preferred SAMv3 approach: sending via UDP socket to port 7655
// rather than using the DATAGRAM SEND command on the SAM bridge socket.
// It blocks until the datagram is sent or an error occurs, respecting the configured timeout.
// Example usage: err := writer.SendDatagram([]byte("hello world"), destinationAddr)
func (w *DatagramWriter) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error {
	if err := w.validateSessionState(); err != nil {
		return err
	}

	logger := w.createSendLogger(dest, len(data))
	logger.Debug("Sending datagram via UDP socket")

	udpConn, err := w.establishUDPConnection(logger)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	udpMessage := w.buildUDPMessage(data, dest, logger)

	if err := w.transmitUDPDatagram(udpConn, udpMessage, logger); err != nil {
		return err
	}

	logger.Debug("Successfully sent datagram via UDP")
	return nil
}

// validateSessionState checks if the session is in a valid state for sending datagrams.
// Returns an error if the session is closed, otherwise returns nil.
func (w *DatagramWriter) validateSessionState() error {
	w.session.mu.RLock()
	defer w.session.mu.RUnlock()

	if w.session.closed {
		return oops.Errorf("session is closed")
	}
	return nil
}

// createSendLogger creates a structured logger with session and destination context.
// The logger includes session ID, destination address, and data size for debugging.
func (w *DatagramWriter) createSendLogger(dest i2pkeys.I2PAddr, dataSize int) *logger.Entry {
	return log.WithFields(logger.Fields{
		"session_id":  w.session.ID(),
		"destination": dest.Base32(),
		"size":        dataSize,
	})
}

// establishUDPConnection creates a UDP connection to the SAM bridge for datagram transmission.
// Uses the configured SAM host or defaults to 127.0.0.1:7655 if not specified.
// Returns the UDP connection or an error if connection establishment fails.
func (w *DatagramWriter) establishUDPConnection(logger *logger.Entry) (*net.UDPConn, error) {
	samHost := w.session.sam.SAMEmit.I2PConfig.SamHost
	if samHost == "" {
		samHost = "127.0.0.1"
	}
	samUDPPort := "7655"

	udpAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(samHost, samUDPPort))
	if err != nil {
		logger.WithError(err).Error("Failed to resolve SAM UDP address")
		return nil, oops.Errorf("failed to resolve SAM UDP address: %w", err)
	}

	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to SAM UDP port")
		return nil, oops.Errorf("failed to connect to SAM UDP port: %w", err)
	}

	return udpConn, nil
}

// buildUDPMessage constructs a SAMv3-compliant UDP datagram message.
// Format: "3.3 <session_id> <destination>\n<data>"
// The header line contains protocol version, session ID, and base64-encoded destination.
func (w *DatagramWriter) buildUDPMessage(data []byte, dest i2pkeys.I2PAddr, log *logger.Entry) []byte {
	sessionID := w.session.ID()
	destination := dest.Base64()

	headerLine := fmt.Sprintf("3.3 %s %s\n", sessionID, destination)
	udpMessage := append([]byte(headerLine), data...)

	log.WithFields(logger.Fields{
		"header":     headerLine,
		"total_size": len(udpMessage),
	}).Debug("Sending UDP datagram to SAM")

	return udpMessage
}

// transmitUDPDatagram sends the constructed UDP message through the connection.
// Writes the complete message to the UDP socket and returns any transmission errors.
func (w *DatagramWriter) transmitUDPDatagram(udpConn *net.UDPConn, udpMessage []byte, logger *logger.Entry) error {
	_, err := udpConn.Write(udpMessage)
	if err != nil {
		logger.WithError(err).Error("Failed to send UDP datagram to SAM")
		return oops.Errorf("failed to send UDP datagram to SAM: %w", err)
	}
	return nil
}
