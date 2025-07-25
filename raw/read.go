package raw

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

// ReceiveDatagram receives a raw datagram from any source
func (r *RawReader) ReceiveDatagram() (*RawDatagram, error) {
	// Check if closed first, but don't rely on this check for safety
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return nil, oops.Errorf("reader is closed")
	}
	r.mu.RUnlock()

	// Use select with closeChan to handle concurrent close operations safely
	select {
	case datagram := <-r.recvChan:
		return datagram, nil
	case err := <-r.errorChan:
		return nil, err
	case <-r.closeChan:
		return nil, oops.Errorf("reader is closed")
	}
}

// Close closes the RawReader and stops its receive loop, cleaning up all associated resources.
// This method is safe to call multiple times and will not block if the reader is already closed.
func (r *RawReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	// Safe session ID retrieval with nil checks for logging
	sessionID := "unknown"
	if r.session != nil && r.session.BaseSession != nil {
		sessionID = r.session.ID()
	}
	logger := log.WithField("session_id", sessionID)
	logger.Debug("Closing RawReader")

	r.closed = true

	// Set atomic flag to indicate we're closing
	atomic.StoreInt32(&r.closing, 1)

	// Signal termination to receiveLoop
	close(r.closeChan)

	// Wait for receiveLoop to signal it has exited
	select {
	case <-r.doneChan:
		// receiveLoop has confirmed it stopped
	case <-time.After(5 * time.Second):
		// Timeout protection - log warning but continue cleanup
		logger.Warn("Timeout waiting for receive loop to stop")
	}

	// Don't close doneChan - let the sender close it

	// Close receiver channels here under mutex protection
	close(r.recvChan)
	close(r.errorChan)

	logger.Debug("Successfully closed RawReader")
	return nil
}

// receiveLoop continuously receives incoming raw datagrams in a separate goroutine.
// This method handles the SAM protocol communication, parses RAW RECEIVED responses,
// and forwards datagrams to the appropriate channels until the reader is closed.
func (r *RawReader) receiveLoop() {
	logger := r.initializeReceiveLoop()
	defer r.signalReceiveLoopCompletion()

	if !r.validateSessionForReceive(logger) {
		return
	}

	r.runMainReceiveLoop(logger)
}

// initializeReceiveLoop sets up logging and returns a configured logger for the receive loop.
func (r *RawReader) initializeReceiveLoop() *logger.Entry {
	sessionID := "unknown"
	if r.session != nil && r.session.BaseSession != nil {
		sessionID = r.session.ID()
	}
	logger := log.WithField("session_id", sessionID)
	logger.Debug("Starting raw receive loop")
	return logger
}

// signalReceiveLoopCompletion signals that the receive loop has completed execution.
func (r *RawReader) signalReceiveLoopCompletion() {
	// Close doneChan to signal completion - channels should be closed by sender
	close(r.doneChan)
}

// validateSessionForReceive checks if the session is valid for receiving operations.
func (r *RawReader) validateSessionForReceive(logger *logger.Entry) bool {
	if r.session == nil {
		logger.Debug("Raw receive loop terminated - session is nil")
		return false
	}

	r.session.mu.RLock()
	defer r.session.mu.RUnlock()

	if r.session.closed || r.session.BaseSession == nil {
		logger.Debug("Raw receive loop terminated - session invalid")
		return false
	}

	return true
}

// runMainReceiveLoop executes the main receive loop that processes datagrams until closed.
func (r *RawReader) runMainReceiveLoop(logger *logger.Entry) {
	for {
		if r.checkForClosure(logger) {
			return
		}

		datagram, err := r.receiveDatagram()
		if err != nil {
			if r.handleReceiveError(err, logger) {
				return
			}
			continue
		}

		if r.forwardDatagram(datagram, logger) {
			return
		}
	}
}

// checkForClosure checks if the reader has been closed in a non-blocking way.
func (r *RawReader) checkForClosure(logger *logger.Entry) bool {
	select {
	case <-r.closeChan:
		logger.Debug("Raw receive loop terminated - reader closed")
		return true
	default:
		return false
	}
}

// handleReceiveError handles errors that occur during datagram reception.
func (r *RawReader) handleReceiveError(err error, logger *logger.Entry) bool {
	select {
	case r.errorChan <- err:
		logger.WithError(err).Error("Failed to receive raw datagram")
	case <-r.closeChan:
		// Reader was closed during error handling
		return true
	}
	return false
}

// forwardDatagram forwards a received datagram to the receive channel.
func (r *RawReader) forwardDatagram(datagram *RawDatagram, logger *logger.Entry) bool {
	// Check atomic flag first to avoid race condition
	if atomic.LoadInt32(&r.closing) == 1 {
		return true
	}

	select {
	case r.recvChan <- datagram:
		logger.Debug("Successfully received raw datagram")
	case <-r.closeChan:
		// Reader was closed during datagram send
		return true
	}
	return false
}

// receiveDatagram handles the low-level protocol parsing for incoming raw datagrams.
// It reads from the SAM connection, parses the RAW RECEIVED response format,
// and constructs RawDatagram objects with decoded data and address information.
func (r *RawReader) receiveDatagram() (*RawDatagram, error) {
	logger := log.WithField("session_id", r.session.ID())

	// Validate session state before processing
	if err := r.validateSessionState(); err != nil {
		return nil, err
	}

	// Read raw response from SAM connection
	response, err := r.readRawResponse()
	if err != nil {
		return nil, err
	}

	logger.WithField("response", response).Debug("Received raw datagram data")

	// Parse the RAW RECEIVED response to extract source and data
	source, data, err := r.parseRawResponse(response)
	if err != nil {
		return nil, err
	}

	// Create and return the raw datagram
	return r.createRawDatagram(source, data)
}

// validateSessionState checks if the session is valid and ready for use.
func (r *RawReader) validateSessionState() error {
	r.session.mu.RLock()
	defer r.session.mu.RUnlock()

	if r.session.closed {
		return oops.Errorf("session is closed")
	}

	if r.session.BaseSession == nil {
		return oops.Errorf("session is not properly initialized")
	}

	if r.session.BaseSession.Conn() == nil {
		return oops.Errorf("session connection is not available")
	}

	return nil
}

// readRawResponse reads the raw response from the SAM connection.
func (r *RawReader) readRawResponse() (string, error) {
	buf := make([]byte, 4096)
	n, err := r.session.Read(buf)
	if err != nil {
		return "", oops.Errorf("failed to read from session: %w", err)
	}

	return string(buf[:n]), nil
}

// parseRawResponse parses the RAW RECEIVED response to extract source and data.
func (r *RawReader) parseRawResponse(response string) (source, data string, err error) {
	scanner := bufio.NewScanner(strings.NewReader(response))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		switch {
		case word == "RAW":
			continue
		case word == "RECEIVED":
			continue
		case strings.HasPrefix(word, "DESTINATION="):
			// Extract source destination from the response
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

	// Validate that we have both source and data
	if source == "" {
		return "", "", oops.Errorf("no source in raw datagram")
	}

	if data == "" {
		return "", "", oops.Errorf("no data in raw datagram")
	}

	return source, data, nil
}

// createRawDatagram creates a RawDatagram from source and data strings.
func (r *RawReader) createRawDatagram(source, data string) (*RawDatagram, error) {
	// Parse the source destination into an I2P address
	sourceAddr, err := i2pkeys.NewI2PAddrFromString(source)
	if err != nil {
		return nil, oops.Errorf("failed to parse source address: %w", err)
	}

	// Decode the base64 data into bytes
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, oops.Errorf("failed to decode raw datagram data: %w", err)
	}

	// Create the raw datagram with decoded data and address information
	datagram := &RawDatagram{
		Data:   decodedData,
		Source: sourceAddr,
		Local:  r.session.Addr(),
	}

	return datagram, nil
}
