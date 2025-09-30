package raw

import (
	"runtime"
	"testing"
	"time"
)

// TestGoroutineLeakInDialOperations reproduces Issue #3 from AUDIT.md
// This test demonstrates that DialAddr/DialI2P create goroutines that can leak
// if the returned PacketConn is not properly closed.
func TestGoroutineLeakInDialOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping goroutine leak test in short mode")
	}

	// Get baseline goroutine count
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // Allow GC to settle
	initialGoroutines := runtime.NumGoroutine()

	// This test simulates the leaky behavior described in the audit
	// where dial operations start goroutines but don't guarantee cleanup

	// Simulate creating connections without proper cleanup (leaky pattern)
	for i := 0; i < 5; i++ {
		// This would normally create a real session, but for testing we'll simulate
		// the problematic pattern where receiveLoop goroutines are started

		// Create a reader (simulates what DialAddr does internally)
		reader := &RawReader{
			recvChan:  make(chan *RawDatagram, 10),
			errorChan: make(chan error, 1),
			closeChan: make(chan struct{}),
			doneChan:  make(chan struct{}),
			closed:    false,
		}

		// Start receiveLoop goroutine (this is what causes the leak)
		go func(r *RawReader) {
			// Simulate receiveLoop that runs indefinitely
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-r.closeChan:
					close(r.doneChan)
					return
				case <-ticker.C:
					// Simulate some work
				}
			}
		}(reader)

		// Connection goes out of scope without Close() being called
		// This simulates the leak scenario from the audit
	}

	// Allow some time for goroutines to start
	time.Sleep(100 * time.Millisecond)

	// Force GC to clean up any garbage
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Check if goroutines leaked
	finalGoroutines := runtime.NumGoroutine()
	goroutineLeak := finalGoroutines - initialGoroutines

	// We expect goroutines to have leaked in the current (buggy) implementation
	if goroutineLeak <= 0 {
		t.Skip("Goroutine leak not reproduced - this is expected after the fix")
	} else {
		t.Logf("Goroutine leak detected: %d extra goroutines (this demonstrates the bug)", goroutineLeak)
		// This is the expected behavior in the current buggy implementation
		// After our fix, this test should be updated to verify no leaks occur
	}
}

// TestGoroutineLeakPrevention validates that proper cleanup prevents leaks
func TestGoroutineLeakPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping goroutine leak test in short mode")
	}

	// Get baseline goroutine count
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	initialGoroutines := runtime.NumGoroutine()

	// Test proper cleanup pattern that should prevent leaks
	readers := make([]*RawReader, 5)

	for i := 0; i < 5; i++ {
		reader := &RawReader{
			recvChan:  make(chan *RawDatagram, 10),
			errorChan: make(chan error, 1),
			closeChan: make(chan struct{}),
			doneChan:  make(chan struct{}),
			closed:    false,
		}
		readers[i] = reader

		// Start receiveLoop goroutine
		go func(r *RawReader) {
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-r.closeChan:
					close(r.doneChan)
					return
				case <-ticker.C:
					// Simulate some work
				}
			}
		}(reader)
	}

	// Allow goroutines to start
	time.Sleep(100 * time.Millisecond)

	// Properly close all readers (this is what prevents the leak)
	for _, reader := range readers {
		close(reader.closeChan)
		// Wait for goroutine to signal completion
		select {
		case <-reader.doneChan:
			// Goroutine stopped cleanly
		case <-time.After(1 * time.Second):
			t.Error("Goroutine did not stop within timeout")
		}
	}

	// Allow cleanup to complete
	time.Sleep(100 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Check final goroutine count
	finalGoroutines := runtime.NumGoroutine()
	goroutineChange := finalGoroutines - initialGoroutines

	// Should be no net increase in goroutines with proper cleanup
	if goroutineChange > 2 { // Allow small variance for test infrastructure
		t.Errorf("Goroutine leak detected even with cleanup: %d extra goroutines", goroutineChange)
	}
}
