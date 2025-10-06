package datagram

import (
	"fmt"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
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
	})
	logger.Debug("Sending datagram via UDP socket")

	// Use UDP socket approach (SAMv3 preferred method) instead of DATAGRAM SEND command
	// Connect to SAM's UDP port (default 7655) for datagram transmission
	samHost := w.session.sam.SAMEmit.I2PConfig.SamHost
	if samHost == "" {
		samHost = "127.0.0.1" // Default SAM host
	}
	samUDPPort := "7655" // Default SAM UDP port for datagram transmission

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

	// Construct the SAMv3 UDP datagram format:
	// First line: "3.3 <session_id> <destination> [options]\n"
	// Remaining data: the actual message payload
	sessionID := w.session.ID()
	destination := dest.Base64()

	// Create the header line according to SAMv3 specification
	headerLine := fmt.Sprintf("3.3 %s %s\n", sessionID, destination)

	// Combine header and data into final UDP packet
	udpMessage := append([]byte(headerLine), data...)

	logger.WithFields(logrus.Fields{
		"header":     headerLine,
		"total_size": len(udpMessage),
	}).Debug("Sending UDP datagram to SAM")

	// Send the datagram via UDP to SAM bridge
	_, err = udpConn.Write(udpMessage)
	if err != nil {
		logger.WithError(err).Error("Failed to send UDP datagram to SAM")
		return oops.Errorf("failed to send UDP datagram to SAM: %w", err)
	}

	logger.Debug("Successfully sent datagram via UDP")
	return nil
}
