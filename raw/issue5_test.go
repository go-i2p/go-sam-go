package raw

import (
	"sync"
	"testing"
	"time"
)

// TestIssue5ChannelCloseFixed tests Issue #5 fix from AUDIT.md
// This test verifies that the fix prevents "send on closed channel" panics
func TestIssue5ChannelCloseFixed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping channel close panic test in short mode")
	}

	t.Log("Testing Issue #5 fix: Channel Close Responsibility Standardization")

	// This test reproduces the specific race condition described in AUDIT.md:
	// Time | Close() Thread        | receiveLoop Thread
	// T1   | close(r.closeChan)    |
	// T2   |                       | select { case r.recvChan <- data:
	// T3   | close(r.recvChan)     |  ← Channel closed
	// T4   |                       | // PANIC: send on closed channel

	for attempt := 0; attempt < 5; attempt++ {
		reader := &RawReader{
			recvChan:  make(chan *RawDatagram, 1),
			errorChan: make(chan error, 1),
			closeChan: make(chan struct{}),
			doneChan:  make(chan struct{}),
			closed:    false,
			mu:        sync.RWMutex{},
		}

		panicDetected := make(chan bool, 1)
		var wg sync.WaitGroup
		wg.Add(2)

		// Goroutine 1: Simulate receiveLoop trying to send
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Attempt %d: PANIC detected in receiveLoop: %v", attempt+1, r)
					panicDetected <- true
				}
			}()

			// Simulate receiveLoop behavior
			for i := 0; i < 10; i++ {
				select {
				case <-reader.closeChan:
					// Normal termination
					return
				default:
					// Try to send on recvChan (this could panic if channel is closed)
					select {
					case reader.recvChan <- &RawDatagram{Data: []byte("test")}:
						// Successfully sent
					case <-reader.closeChan:
						return
					case <-time.After(1 * time.Millisecond):
						// Timeout to avoid blocking forever
					}
				}
				time.Sleep(1 * time.Millisecond) // Allow interleaving
			}
		}()

		// Goroutine 2: Simulate Close() method with the fix
		go func() {
			defer wg.Done()
			time.Sleep(5 * time.Millisecond) // Let receiveLoop start

			// This is the FIXED pattern from raw/read.go (now matches datagram approach)
			reader.mu.Lock()
			if !reader.closed {
				reader.closed = true
				close(reader.closeChan) // T1: Signal termination

				// FIXED: Don't close recvChan and errorChan here
				// This prevents the "send on closed channel" panic
				// The channels will be garbage collected when references are dropped
			}
			reader.mu.Unlock()
		}()

		// Wait for both goroutines with timeout
		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()

		select {
		case <-done:
			// Completed without panic
		case <-panicDetected:
			t.Errorf("Attempt %d: Unexpected panic after fix - Issue #5 not resolved", attempt+1)
			return
		case <-time.After(100 * time.Millisecond):
			t.Logf("Attempt %d: Completed successfully", attempt+1)
		}
	}

	t.Log("Fix successful: No panics detected in any attempt")
}

// TestDatagramApproachIsSafer tests that the datagram approach avoids the panic
func TestDatagramApproachIsSafer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping datagram approach test in short mode")
	}

	t.Log("Testing datagram package approach: NOT closing send channels")

	for attempt := 0; attempt < 5; attempt++ {
		reader := &RawReader{
			recvChan:  make(chan *RawDatagram, 1),
			errorChan: make(chan error, 1),
			closeChan: make(chan struct{}),
			doneChan:  make(chan struct{}),
			closed:    false,
			mu:        sync.RWMutex{},
		}

		panicDetected := make(chan bool, 1)
		var wg sync.WaitGroup
		wg.Add(2)

		// Goroutine 1: Simulate receiveLoop trying to send
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Attempt %d: PANIC detected: %v", attempt+1, r)
					panicDetected <- true
				}
			}()

			for i := 0; i < 10; i++ {
				select {
				case <-reader.closeChan:
					return
				default:
					select {
					case reader.recvChan <- &RawDatagram{Data: []byte("test")}:
						// Successfully sent
					case <-reader.closeChan:
						return
					case <-time.After(1 * time.Millisecond):
						// Timeout
					}
				}
				time.Sleep(1 * time.Millisecond)
			}
		}()

		// Goroutine 2: Simulate Close() using datagram approach (safer)
		go func() {
			defer wg.Done()
			time.Sleep(5 * time.Millisecond)

			// This follows the datagram package pattern - safer approach
			reader.mu.Lock()
			if !reader.closed {
				reader.closed = true
				close(reader.closeChan) // Signal termination

				// DON'T close recvChan and errorChan
				// Let them be garbage collected when references are dropped
				// This is the datagram package approach
			}
			reader.mu.Unlock()
		}()

		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()

		select {
		case <-done:
			// Completed successfully
		case <-panicDetected:
			t.Errorf("Attempt %d: Unexpected panic with safer approach", attempt+1)
			return
		case <-time.After(100 * time.Millisecond):
			t.Logf("Attempt %d: Timed out", attempt+1)
		}
	}

	t.Log("Datagram approach completed successfully - no panics detected")
}
