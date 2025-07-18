package datagram

import (
	"bufio"
	"encoding/base64"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/go-i2p/logger"
	"github.com/samber/oops"
)

// ReceiveDatagram receives a single datagram from the I2P network.
// This method blocks until a datagram is received or an error occurs, returning
// the received datagram with its data and addressing information. It handles
// concurrent access safely and provides proper error handling for network issues.
// Example usage: datagram, err := reader.ReceiveDatagram()
func (r *DatagramReader) ReceiveDatagram() (*Datagram, error) {
	// Check if the reader is closed before attempting to receive
	// This prevents operations on invalid readers and provides clear error messages
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return nil, oops.Errorf("reader is closed")
	}
	r.mu.RUnlock()

	// Use select to handle multiple channel operations atomically
	// This ensures proper handling of datagrams, errors, and close signals
	select {
	case datagram := <-r.recvChan:
		// Successfully received a datagram from the network
		return datagram, nil
	case err := <-r.errorChan:
		// An error occurred during datagram reception
		return nil, err
	case <-r.closeChan:
		// The reader has been closed while waiting for a datagram
		return nil, oops.Errorf("reader is closed")
	}
}

// Close closes the DatagramReader and stops its receive loop.
// This method safely terminates the reader, cleans up all associated resources,
// and signals any waiting goroutines to stop. It's safe to call multiple times
// and will not block if the reader is already closed.
// Example usage: defer reader.Close()
func (r *DatagramReader) Close() error {
	// Use sync.Once to ensure cleanup only happens once
	// This prevents double-close panics and ensures thread safety
	r.closeOnce.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.closed {
			return
		}

		// Log reader closure for debugging and monitoring
		sessionID := "unknown"
		if r.session != nil && r.session.BaseSession != nil {
			sessionID = r.session.ID()
		}
		logger := log.WithField("session_id", sessionID)
		logger.Debug("Closing DatagramReader")

		r.closed = true

		// Set atomic flag to indicate we're closing
		atomic.StoreInt32(&r.closing, 1)

		// Signal the receive loop to terminate
		// This prevents the background goroutine from continuing to run
		close(r.closeChan)

		// Wait for the receive loop to confirm termination
		// This ensures proper cleanup before returning
		select {
		case <-r.doneChan:
			// Receive loop has confirmed it stopped
			logger.Debug("Receive loop stopped")
		case <-time.After(5 * time.Second):
			// Timeout protection to prevent indefinite blocking
			logger.Warn("Timeout waiting for receive loop to stop")
		}

		// Clean up channels to prevent resource leaks
		// Close channels that are safe to close
		close(r.recvChan)
		close(r.errorChan)

		logger.Debug("Successfully closed DatagramReader")
	})

	return nil
}

// receiveLoop continuously receives incoming datagrams in a separate goroutine.
// This method handles the SAM protocol communication for datagram reception, parsing
// DATAGRAM RECEIVED responses and forwarding datagrams to the appropriate channels.
// It runs until the reader is closed and provides error handling for network issues.
func (r *DatagramReader) receiveLoop() {
	logger := r.initializeReceiveLoop()
	defer r.signalReceiveLoopCompletion()

	if err := r.validateSessionState(logger); err != nil {
		return
	}

	r.runReceiveLoop(logger)
}

// initializeReceiveLoop sets up logging context and returns a logger for the receive loop.
func (r *DatagramReader) initializeReceiveLoop() *logger.Entry {
	sessionID := "unknown"
	if r.session != nil && r.session.BaseSession != nil {
		sessionID = r.session.ID()
	}
	logger := log.WithField("session_id", sessionID)
	logger.Debug("Starting datagram receive loop")
	return logger
}

// signalReceiveLoopCompletion signals that the receive loop has completed execution.
func (r *DatagramReader) signalReceiveLoopCompletion() {
	// Close doneChan to signal completion - channels should be closed by sender
	close(r.doneChan)
}

// validateSessionState checks if the session is valid before starting the receive loop.
func (r *DatagramReader) validateSessionState(logger *logger.Entry) error {
	if r.session == nil || r.session.BaseSession == nil {
		logger.Error("Invalid session state")
		select {
		case r.errorChan <- oops.Errorf("invalid session state"):
		case <-r.closeChan:
		}
		return oops.Errorf("invalid session state")
	}
	return nil
}

// runReceiveLoop executes the main receive loop until the reader is closed.
func (r *DatagramReader) runReceiveLoop(logger *logger.Entry) {
	for {
		select {
		case <-r.closeChan:
			logger.Debug("Receive loop terminated")
			return
		default:
			if !r.processIncomingDatagram(logger) {
				return
			}
		}
	}
}

// processIncomingDatagram receives and forwards a single datagram, returning false if the loop should terminate.
func (r *DatagramReader) processIncomingDatagram(logger *logger.Entry) bool {
	// Check atomic flag first to avoid race condition
	if atomic.LoadInt32(&r.closing) == 1 {
		return false
	}

	datagram, err := r.receiveDatagram()
	if err != nil {
		logger.WithError(err).Debug("Error receiving datagram")
		select {
		case r.errorChan <- err:
		case <-r.closeChan:
			return false
		}
		return true
	}

	// Check atomic flag again before sending to avoid race condition
	if atomic.LoadInt32(&r.closing) == 1 {
		return false
	}

	select {
	case r.recvChan <- datagram:
		// Successfully forwarded the datagram
	case <-r.closeChan:
		return false
	}
	return true
}

// receiveDatagram performs the actual datagram reception from the SAM bridge.
// This method handles the low-level SAM protocol communication, parsing DATAGRAM RECEIVED
// responses and extracting the datagram data and addressing information. It provides
// the core functionality for the receive loop and handles protocol-specific details.
// receiveDatagram performs the actual datagram reception from the SAM bridge
func (r *DatagramReader) receiveDatagram() (*Datagram, error) {
	// Read data from the SAM connection
	// This blocks until data is available or an error occurs
	conn := r.session.Conn()
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, oops.Errorf("failed to read from SAM connection: %w", err)
	}

	// Parse the received data as a SAM protocol message
	// The message format follows SAMv3 specifications for datagram reception
	response := string(buffer[:n])
	log.WithField("response", response).Debug("Received SAM response")

	// Parse the response to extract datagram information
	// This involves parsing the SAM protocol format and extracting the payload
	if !strings.Contains(response, "DATAGRAM RECEIVED") {
		return nil, oops.Errorf("unexpected response format: %s", response)
	}

	// Parse the DATAGRAM RECEIVED response
	scanner := bufio.NewScanner(strings.NewReader(response))
	scanner.Split(bufio.ScanWords)

	var source, data string
	for scanner.Scan() {
		word := scanner.Text()
		switch {
		case word == "DATAGRAM":
			continue
		case word == "RECEIVED":
			continue
		case strings.HasPrefix(word, "DESTINATION="):
			source = word[12:]
		case strings.HasPrefix(word, "SIZE="):
			continue // We'll get the actual data size from the payload
		default:
			// Remaining data is the base64-encoded payload
			if data == "" {
				data = word
			} else {
				data += " " + word
			}
		}
	}

	if source == "" {
		return nil, oops.Errorf("no source in datagram")
	}

	if data == "" {
		return nil, oops.Errorf("no data in datagram")
	}

	// Parse the source destination
	sourceAddr, err := i2pkeys.NewI2PAddrFromString(source)
	if err != nil {
		return nil, oops.Errorf("failed to parse source address: %w", err)
	}

	// Decode the base64 data
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, oops.Errorf("failed to decode datagram data: %w", err)
	}

	// Create the datagram
	datagram := &Datagram{
		Data:   decodedData,
		Source: sourceAddr,
		Local:  r.session.Addr(),
	}

	return datagram, nil
}
