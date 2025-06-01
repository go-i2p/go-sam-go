package datagram

import (
	"bufio"
	"encoding/base64"
	"strings"
	"time"

	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
)

// ReceiveDatagram receives a datagram from any source
func (r *DatagramReader) ReceiveDatagram() (*Datagram, error) {
	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return nil, oops.Errorf("reader is closed")
	}
	r.mu.RUnlock()

	select {
	case datagram := <-r.recvChan:
		return datagram, nil
	case err := <-r.errorChan:
		return nil, err
	case <-r.closeChan:
		return nil, oops.Errorf("reader is closed")
	}
}

func (r *DatagramReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	logger := log.WithField("session_id", r.session.ID())
	logger.Debug("Closing DatagramReader")

	r.closed = true

	// Signal termination to receiveLoop
	close(r.closeChan)

	// Wait for receiveLoop to signal completion with timeout protection
	select {
	case <-r.doneChan:
		// receiveLoop has confirmed it stopped
		logger.Debug("Receive loop terminated gracefully")
	case <-time.After(5 * time.Second):
		// Timeout protection - log warning but continue cleanup
		logger.Warn("Timeout waiting for receive loop to stop")
	}

	// Close receiver channels only after receiveLoop has stopped
	// Use non-blocking close to avoid deadlock if channels are already closed
	r.safeCloseChannel()

	logger.Debug("Successfully closed DatagramReader")
	return nil
}

// safeCloseChannel safely closes channels with panic recovery
func (r *DatagramReader) safeCloseChannel() {
	defer func() {
		if recover() != nil {
			// Channel was already closed - this is expected in some race conditions
		}
	}()

	close(r.recvChan)
	close(r.errorChan)
}

// receiveLoop continuously receives incoming datagrams
func (r *DatagramReader) receiveLoop() {
	logger := log.WithField("session_id", r.session.ID())
	logger.Debug("Starting receive loop")

	// Ensure doneChan is properly signaled when loop exits
	defer func() {
		// Use non-blocking send to avoid deadlock if Close() isn't waiting
		select {
		case r.doneChan <- struct{}{}:
		default:
		}
		logger.Debug("Receive loop goroutine terminated")
	}()

	for {
		// Check for closure signal before any blocking operations
		select {
		case <-r.closeChan:
			logger.Debug("Receive loop terminated - reader closed")
			return
		default:
		}

		// Perform the blocking read operation
		datagram, err := r.receiveDatagram()
		if err != nil {
			// Use atomic send pattern with close check to prevent panic
			select {
			case r.errorChan <- err:
				logger.WithError(err).Error("Failed to receive datagram")
			case <-r.closeChan:
				// Reader was closed during error handling - exit gracefully
				logger.Debug("Receive loop terminated during error handling")
				return
			}
			continue
		}

		// Send the datagram or handle closure atomically
		select {
		case r.recvChan <- datagram:
			logger.Debug("Successfully received datagram")
		case <-r.closeChan:
			// Reader was closed during datagram send - exit gracefully
			logger.Debug("Receive loop terminated during datagram send")
			return
		}
	}
}

// receiveDatagram handles the low-level datagram reception
func (r *DatagramReader) receiveDatagram() (*Datagram, error) {
	logger := log.WithField("session_id", r.session.ID())

	// Read from the session connection for incoming datagrams
	buf := make([]byte, 4096)
	n, err := r.session.Read(buf)
	if err != nil {
		return nil, oops.Errorf("failed to read from session: %w", err)
	}

	response := string(buf[:n])
	logger.WithField("response", response).Debug("Received datagram data")

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
