package raw

import (
	"bufio"
	"encoding/base64"
	"strings"
	"time"

	"github.com/go-i2p/i2pkeys"
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

func (r *RawReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	logger := log.WithField("session_id", r.session.ID())
	logger.Debug("Closing RawReader")

	r.closed = true

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

	// Fix: Close doneChan here to prevent multiple closes
	close(r.doneChan)

	// Fix: Close receiver channels here under mutex protection
	close(r.recvChan)
	close(r.errorChan)

	logger.Debug("Successfully closed RawReader")
	return nil
}

// receiveLoop continuously receives incoming raw datagrams
func (r *RawReader) receiveLoop() {
	logger := log.WithField("session_id", r.session.ID())
	logger.Debug("Starting raw receive loop")

	// Signal completion when this loop exits
	defer func() {
		select {
		case r.doneChan <- struct{}{}:
			// Successfully signaled completion
		default:
			// Channel may be closed or blocked - that's okay
		}
	}()

	// Check session state before starting loop
	r.session.mu.RLock()
	if r.session.closed || r.session.BaseSession == nil {
		r.session.mu.RUnlock()
		logger.Debug("Raw receive loop terminated - session invalid")
		return
	}
	r.session.mu.RUnlock()

	for {
		// Check for closure in a non-blocking way first
		select {
		case <-r.closeChan:
			logger.Debug("Raw receive loop terminated - reader closed")
			return
		default:
		}

		// Now perform the blocking read operation
		datagram, err := r.receiveDatagram()
		if err != nil {
			// Use atomic check and send pattern to avoid race
			select {
			case r.errorChan <- err:
				logger.WithError(err).Error("Failed to receive raw datagram")
			case <-r.closeChan:
				// Reader was closed during error handling
				return
			}
			continue
		}

		// Send the datagram or handle closure atomically
		select {
		case r.recvChan <- datagram:
			logger.Debug("Successfully received raw datagram")
		case <-r.closeChan:
			// Reader was closed during datagram send
			return
		}
	}
}

// receiveDatagram handles the low-level raw datagram reception
func (r *RawReader) receiveDatagram() (*RawDatagram, error) {
	logger := log.WithField("session_id", r.session.ID())

	// Check if session is valid and not closed
	r.session.mu.RLock()
	if r.session.closed {
		r.session.mu.RUnlock()
		return nil, oops.Errorf("session is closed")
	}

	// Check if BaseSession is properly initialized
	if r.session.BaseSession == nil {
		r.session.mu.RUnlock()
		return nil, oops.Errorf("session is not properly initialized")
	}

	if r.session.BaseSession.Conn() == nil {
		r.session.mu.RUnlock()
		return nil, oops.Errorf("session connection is not available")
	}
	r.session.mu.RUnlock()

	// Read from the session connection for incoming raw datagrams
	buf := make([]byte, 4096)
	n, err := r.session.Read(buf)
	if err != nil {
		return nil, oops.Errorf("failed to read from session: %w", err)
	}

	response := string(buf[:n])
	logger.WithField("response", response).Debug("Received raw datagram data")

	// Parse the RAW RECEIVED response
	scanner := bufio.NewScanner(strings.NewReader(response))
	scanner.Split(bufio.ScanWords)

	var source, data string
	for scanner.Scan() {
		word := scanner.Text()
		switch {
		case word == "RAW":
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
		return nil, oops.Errorf("no source in raw datagram")
	}

	if data == "" {
		return nil, oops.Errorf("no data in raw datagram")
	}

	// Parse the source destination
	sourceAddr, err := i2pkeys.NewI2PAddrFromString(source)
	if err != nil {
		return nil, oops.Errorf("failed to parse source address: %w", err)
	}

	// Decode the base64 data
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, oops.Errorf("failed to decode raw datagram data: %w", err)
	}

	// Create the raw datagram
	datagram := &RawDatagram{
		Data:   decodedData,
		Source: sourceAddr,
		Local:  r.session.Addr(),
	}

	return datagram, nil
}
