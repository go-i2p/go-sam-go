package raw

import (
	"sync"
	"testing"
	"time"
)

// TestSessionClosureRaceConditionFixed tests that Issue #1 from AUDIT.md is resolved
// This validates that ReceiveDatagram() properly synchronizes with Close() operations
func TestSessionClosureRaceConditionFixed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	// Test that demonstrates the fix works by ensuring clean error handling
	for attempt := 0; attempt < 10; attempt++ {
		reader := &RawReader{
			recvChan:  make(chan *RawDatagram, 1),
			errorChan: make(chan error, 1),
			closeChan: make(chan struct{}),
			doneChan:  make(chan struct{}),
			closed:    false,
			mu:        sync.RWMutex{},
		}

		// Send close signal after a delay
		go func() {
			time.Sleep(10 * time.Millisecond)
			close(reader.closeChan)
		}()

		// This should now cleanly return with "reader is closed" error
		// The fix ensures no race between the closed check and channel operation
		_, err := reader.ReceiveDatagram()

		if err == nil {
			t.Error("Expected error when receiving on closed reader")
		} else if err.Error() != "reader is closed" {
			t.Errorf("Expected 'reader is closed' error, got: %v", err)
		}
	}
}

// TestSessionClosureRaceConditionWithChannels tests proper closure signaling
func TestSessionClosureWithProperSignaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	// Test proper closure pattern that works with the fix
	reader := &RawReader{
		recvChan:  make(chan *RawDatagram, 1),
		errorChan: make(chan error, 1),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),
		closed:    false,
		mu:        sync.RWMutex{},
	}

	// Test 1: Pre-closed reader should return immediately
	reader.mu.Lock()
	reader.closed = true
	reader.mu.Unlock()

	_, err := reader.ReceiveDatagram()
	if err == nil || err.Error() != "reader is closed" {
		t.Errorf("Expected 'reader is closed' error for pre-closed reader, got: %v", err)
	}

	// Test 2: Reader closed via channel signal should work correctly
	reader2 := &RawReader{
		recvChan:  make(chan *RawDatagram, 1),
		errorChan: make(chan error, 1),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),
		closed:    false,
		mu:        sync.RWMutex{},
	}

	// Start receive operation that will wait on channels
	resultChan := make(chan error, 1)
	go func() {
		_, err := reader2.ReceiveDatagram()
		resultChan <- err
	}()

	// Give the goroutine time to start and enter the select
	time.Sleep(10 * time.Millisecond)

	// Close via channel signal - this should wake up the ReceiveDatagram
	close(reader2.closeChan)

	// Wait for result
	select {
	case err := <-resultChan:
		if err == nil || err.Error() != "reader is closed" {
			t.Errorf("Expected 'reader is closed' error, got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("ReceiveDatagram should have returned within 1 second")
	}
}
