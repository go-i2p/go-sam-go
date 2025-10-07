package raw

import (
	"time"

	"github.com/go-i2p/logger"
	"github.com/samber/oops"
)

// ReceiveDatagram receives a raw datagram from any source
func (r *RawReader) ReceiveDatagram() (*RawDatagram, error) {
	// Hold read lock for the entire operation to prevent race with Close()
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return nil, oops.Errorf("reader is closed")
	}

	// Use select with closeChan to handle concurrent close operations safely
	// The lock ensures that channels won't be closed while we're selecting on them
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

	if r.checkRawReaderAlreadyClosed() {
		return nil
	}

	logger := r.initializeRawCloseLogger()
	r.signalRawReaderClosure(logger)
	r.waitForRawReceiveLoopTermination(logger)
	r.finalizeRawReaderClosure(logger)

	return nil
}

// checkRawReaderAlreadyClosed returns true if the reader is already closed.
func (r *RawReader) checkRawReaderAlreadyClosed() bool {
	return r.closed
}

// initializeRawCloseLogger sets up logging context for the raw reader close operation.
func (r *RawReader) initializeRawCloseLogger() *logger.Entry {
	// Safe session ID retrieval with nil checks for logging
	sessionID := "unknown"
	if r.session != nil && r.session.BaseSession != nil {
		sessionID = r.session.ID()
	}
	logger := log.WithField("session_id", sessionID)
	logger.Debug("Closing RawReader")
	return logger
}

// signalRawReaderClosure marks the reader as closed and signals termination.
func (r *RawReader) signalRawReaderClosure(logger *logger.Entry) {
	r.closed = true
	// Signal termination to receiveLoop
	close(r.closeChan)
}

// waitForRawReceiveLoopTermination waits for the receive loop to stop with timeout protection.
func (r *RawReader) waitForRawReceiveLoopTermination(logger *logger.Entry) {
	// Wait for receiveLoop to signal it has exited
	select {
	case <-r.doneChan:
		// receiveLoop has confirmed it stopped
	case <-time.After(1 * time.Second):
		// Shorter timeout for tests - log warning but continue cleanup
		logger.Warn("Timeout waiting for receive loop to stop")
	}
}

// finalizeRawReaderClosure performs final cleanup and logging.
func (r *RawReader) finalizeRawReaderClosure(logger *logger.Entry) {
	// Don't close doneChan - let the sender close it

	// Clean up channels to prevent resource leaks
	// Note: We don't close r.recvChan and r.errorChan here because the receive loop
	// might still be sending on them. These channels will be garbage collected when
	// all references are dropped. Only the receive loop should close send-channels.
	// This matches the datagram package approach and prevents "send on closed channel" panics.

	logger.Debug("Successfully closed RawReader")
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
	// Check closed state with consistent mutex protection
	r.mu.RLock()
	isClosed := r.closed
	r.mu.RUnlock()

	if isClosed {
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

// receiveDatagram handles the low-level protocol for incoming raw datagrams using SAMv3 UDP forwarding.
// V1/V2 TCP control socket reading is no longer supported.
func (r *RawReader) receiveDatagram() (*RawDatagram, error) {
	// Validate session state before processing
	if err := r.validateSessionState(); err != nil {
		return nil, err
	}

	// V3-only: Read from UDP connection
	r.session.mu.RLock()
	udpConn := r.session.udpConn
	r.session.mu.RUnlock()

	if udpConn == nil {
		return nil, oops.Errorf("UDP connection not available (v3 UDP forwarding required)")
	}

	return r.session.readRawFromUDP(udpConn)
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

	return nil
}
