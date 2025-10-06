package raw

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-i2p/common/base64"

	"github.com/go-i2p/i2pkeys"

	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// SetTimeout sets the timeout for raw datagram operations.
// This method configures the maximum time to wait for send operations to complete.
// It returns the writer instance for method chaining.
// Example usage: writer.SetTimeout(30*time.Second).SendDatagram(data, dest)
func (w *RawWriter) SetTimeout(timeout time.Duration) *RawWriter {
	w.timeout = timeout
	return w
}

// SendDatagram sends a raw datagram to the specified destination.
// This method handles the complete send operation including data encoding,
// SAM protocol communication, and response parsing for error handling.
// Example usage: err := writer.SendDatagram([]byte("hello"), destAddr)
func (w *RawWriter) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error {
	w.session.mu.RLock()
	if w.session.closed {
		w.session.mu.RUnlock()
		return oops.Errorf("session is closed")
	}
	w.session.mu.RUnlock()

	logger := log.WithFields(logrus.Fields{
		"session_id":  w.session.ID(),
		"destination": dest.Base32(),
		"size":        len(data),
	})
	logger.Debug("Sending raw datagram")

	// Encode the data as base64 for SAM protocol transmission
	encodedData := base64.I2PEncoding.EncodeToString(data)

	// Create the RAW SEND command following SAMv3 protocol format
	// The command includes session ID, destination, size, and base64-encoded data
	sendCmd := fmt.Sprintf("RAW SEND ID=%s DESTINATION=%s SIZE=%d\n%s\n",
		w.session.ID(), dest.Base64(), len(data), encodedData)

	logger.WithField("command", strings.Split(sendCmd, "\n")[0]).Debug("Sending RAW SEND")

	// Send the command to the SAM bridge over the session connection
	_, err := w.session.Write([]byte(sendCmd))
	if err != nil {
		logger.WithError(err).Error("Failed to send raw datagram")
		return oops.Errorf("failed to send raw datagram: %w", err)
	}

	// Read the response from the SAM bridge to determine send status
	buf := make([]byte, 1024)
	n, err := w.session.Read(buf)
	if err != nil {
		logger.WithError(err).Error("Failed to read send response")
		return oops.Errorf("failed to read send response: %w", err)
	}

	response := string(buf[:n])
	logger.WithField("response", response).Debug("Received send response")

	// Parse the response to check for errors and handle failure conditions
	if err := w.parseSendResponse(response); err != nil {
		return err
	}

	logger.Debug("Successfully sent raw datagram")
	return nil
}

// parseSendResponse parses the RAW STATUS response from the SAM bridge after sending a datagram.
// It examines the response string to determine if the send operation was successful or failed,
// and returns appropriate error messages for different failure conditions like unreachable peers,
// invalid keys, timeouts, and other I2P network errors.
// Example response: "RAW STATUS RESULT=OK" or "RAW STATUS RESULT=CANT_REACH_PEER"
func (w *RawWriter) parseSendResponse(response string) error {
	// Check for successful send operation first
	if strings.Contains(response, "RESULT=OK") {
		return nil
	}

	// Handle specific error conditions returned by the SAM bridge
	// These errors provide meaningful feedback about I2P network failures
	switch {
	case strings.Contains(response, "RESULT=CANT_REACH_PEER"):
		return oops.Errorf("cannot reach peer")
	case strings.Contains(response, "RESULT=I2P_ERROR"):
		return oops.Errorf("I2P internal error")
	case strings.Contains(response, "RESULT=INVALID_KEY"):
		return oops.Errorf("invalid destination key")
	case strings.Contains(response, "RESULT=INVALID_ID"):
		return oops.Errorf("invalid session ID")
	case strings.Contains(response, "RESULT=TIMEOUT"):
		return oops.Errorf("send timeout")
	default:
		// Handle unknown error responses by extracting the result portion
		if strings.HasPrefix(response, "RAW STATUS RESULT=") {
			result := strings.TrimSpace(response[18:])
			return oops.Errorf("send failed: %s", result)
		}
		return oops.Errorf("unexpected response format: %s", response)
	}
}
