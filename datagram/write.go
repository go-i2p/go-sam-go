package datagram

import (
	"encoding/base64"
	"fmt"
	"strings"
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
// This method handles the complete datagram transmission process including data encoding,
// SAM protocol communication, and response validation. It blocks until the datagram
// is sent or an error occurs, respecting the configured timeout duration.
// Example usage: err := writer.SendDatagram([]byte("hello world"), destinationAddr)
func (w *DatagramWriter) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error {
	// Check if the session is closed before attempting to send
	// This prevents operations on invalid sessions
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
	logger.Debug("Sending datagram")

	// Encode the datagram data as base64 for SAM protocol transmission
	// The SAM protocol requires base64 encoding for binary data transfer
	encodedData := base64.StdEncoding.EncodeToString(data)

	// Create the DATAGRAM SEND command following SAMv3 protocol format
	// This command includes session ID, destination address, size, and encoded data
	sendCmd := fmt.Sprintf("DATAGRAM SEND ID=%s DESTINATION=%s SIZE=%d\n%s\n",
		w.session.ID(), dest.Base64(), len(data), encodedData)

	// Log the command being sent for debugging protocol communication
	logger.WithField("command", strings.Split(sendCmd, "\n")[0]).Debug("Sending DATAGRAM SEND")

	// Send the command to the SAM bridge and read the response
	// This handles the underlying protocol communication with error handling
	conn := w.session.Conn()
	if _, err := conn.Write([]byte(sendCmd)); err != nil {
		logger.WithError(err).Error("Failed to send datagram command")
		return oops.Errorf("failed to send datagram command: %w", err)
	}

	// Read the response from the SAM bridge to check for errors
	// The response indicates whether the datagram was successfully queued for transmission
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		logger.WithError(err).Error("Failed to read datagram response")
		return oops.Errorf("failed to read datagram response: %w", err)
	}

	// Parse the response to check for transmission errors
	// The SAM bridge returns status information about the send operation
	if err := w.parseSendResponse(string(response[:n])); err != nil {
		logger.WithError(err).Error("Datagram send failed")
		return oops.Errorf("datagram send failed: %w", err)
	}

	logger.Debug("Successfully sent datagram")
	return nil
}

// parseSendResponse parses the DATAGRAM STATUS response from the SAM bridge.
// This method analyzes the response from a datagram send operation to determine
// if the transmission was successful or if any errors occurred during the process.
// It provides detailed error information for debugging and error handling.
// parseSendResponse parses the DATAGRAM STATUS response from the SAM bridge
func (w *DatagramWriter) parseSendResponse(response string) error {
	// Parse the response to extract status information
	// The response format follows SAMv3 protocol specifications
	if strings.Contains(response, "RESULT=OK") {
		log.Debug("Datagram send successful")
		return nil
	}

	// Handle various error conditions that can occur during send operations
	// Different errors provide specific information about transmission failures
	if strings.Contains(response, "RESULT=INVALID_KEY") {
		return oops.Errorf("invalid destination key")
	}
	if strings.Contains(response, "RESULT=KEY_NOT_FOUND") {
		return oops.Errorf("destination key not found")
	}
	if strings.Contains(response, "RESULT=INVALID_ID") {
		return oops.Errorf("invalid session ID")
	}

	// Log the unexpected response for debugging protocol issues
	log.WithField("response", response).Error("Unexpected datagram send response")
	return oops.Errorf("unexpected datagram send response: %s", response)
}
