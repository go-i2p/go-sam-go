package datagram2

import (
	"fmt"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/go-i2p/logger"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// SetTimeout sets the timeout for datagram2 write operations.
// This method configures the maximum time to wait for authenticated datagram send operations
// to complete. Returns the writer instance for method chaining convenience.
// Example usage: writer.SetTimeout(30*time.Second).SendDatagram(data, destination)
func (w *Datagram2Writer) SetTimeout(timeout time.Duration) *Datagram2Writer {
	// Configure the timeout for send operations to prevent indefinite blocking
	w.timeout = timeout
	return w
}

// SendDatagram sends an authenticated datagram with replay protection to the specified I2P destination.
// It uses the SAMv3 UDP approach by sending to port 7655 with DATAGRAM2 format.
// Maximum datagram size is 31744 bytes (11 KB recommended for reliability).
// Example usage: err := writer.SendDatagram([]byte("hello world"), destinationAddr)
func (w *Datagram2Writer) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error {
	if err := w.validateSessionState(); err != nil {
		return err
	}

	logger := w.createSendLogger(dest, len(data))
	logger.Debug("Sending datagram2 message via UDP socket")

	udpConn, err := w.establishUDPConnection(logger)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	udpMessage := w.buildUDPMessage(data, dest, logger)

	return w.transmitUDPMessage(udpConn, udpMessage, logger)
}

// validateSessionState checks if the session is closed before attempting to send.
// Returns an error if the session is closed, nil otherwise.
func (w *Datagram2Writer) validateSessionState() error {
	w.session.mu.RLock()
	defer w.session.mu.RUnlock()

	if w.session.closed {
		return oops.Errorf("session is closed")
	}
	return nil
}

// createSendLogger creates a logger with contextual fields for send operations.
// Returns a configured logger instance with session ID, destination, size, and style fields.
func (w *Datagram2Writer) createSendLogger(dest i2pkeys.I2PAddr, dataSize int) *logger.Entry {
	return log.WithFields(logrus.Fields{
		"session_id":  w.session.ID(),
		"destination": dest.Base32(),
		"size":        dataSize,
		"style":       "DATAGRAM2",
	})
}

// establishUDPConnection resolves the SAM UDP address and establishes a UDP connection.
// Returns the UDP connection or an error if resolution or connection fails.
func (w *Datagram2Writer) establishUDPConnection(logger *logger.Entry) (*net.UDPConn, error) {
	samHost := w.session.sam.SAMEmit.I2PConfig.SamHost
	if samHost == "" {
		samHost = "127.0.0.1" // Default SAM host
	}
	samUDPPort := "7655" // Default SAM UDP port for datagram2 transmission

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

// buildUDPMessage constructs the SAMv3 UDP datagram2 message format.
// Returns the complete UDP message with header and payload combined.
func (w *Datagram2Writer) buildUDPMessage(data []byte, dest i2pkeys.I2PAddr, logger *logger.Entry) []byte {
	sessionID := w.session.ID()
	destination := dest.Base64()

	// Create the header line according to SAMv3 specification
	// The SAM bridge handles DATAGRAM2 authentication and replay protection internally
	headerLine := fmt.Sprintf("3.3 %s %s\n", sessionID, destination)

	// Combine header and data into final UDP packet
	udpMessage := append([]byte(headerLine), data...)

	logger.WithFields(logrus.Fields{
		"header":     headerLine,
		"total_size": len(udpMessage),
	}).Debug("Sending UDP datagram2 to SAM")

	return udpMessage
}

// transmitUDPMessage sends the constructed UDP message to the SAM bridge.
// Returns an error if the transmission fails, nil otherwise.
func (w *Datagram2Writer) transmitUDPMessage(udpConn *net.UDPConn, udpMessage []byte, logger *logger.Entry) error {
	_, err := udpConn.Write(udpMessage)
	if err != nil {
		logger.WithError(err).Error("Failed to send UDP datagram2 to SAM")
		return oops.Errorf("failed to send UDP datagram2 to SAM: %w", err)
	}

	logger.Debug("Successfully sent datagram2 message via UDP")
	return nil
}
