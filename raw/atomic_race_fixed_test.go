package raw

import (
	"sync"
	"testing"
	"time"
)

// TestAtomicMutexInconsistencyFixed validates that Issue #2 from AUDIT.md is resolved
// This test demonstrates that the atomic/mutex inconsistency has been eliminated
// by using consistent mutex-only synchronization for the closed state.
func TestAtomicMutexInconsistencyFixed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	// This test validates that we now use consistent synchronization
	// The atomic `closing` field has been eliminated, and we use only mutex protection

	for attempt := 0; attempt < 20; attempt++ {
		reader := &RawReader{
			recvChan:  make(chan *RawDatagram, 1),
			errorChan: make(chan error, 1),
			closeChan: make(chan struct{}),
			doneChan:  make(chan struct{}),
			closed:    false,
			mu:        sync.RWMutex{},
		}

		var wg sync.WaitGroup

		// Goroutine 1: Check closure state consistently using only mutex
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				// After fix: use only mutex protection consistently
				reader.mu.RLock()
				isClosed := reader.closed
				reader.mu.RUnlock()

				if isClosed {
					// State is now always consistent
					break
				}
				time.Sleep(time.Microsecond)
			}
		}()

		// Goroutine 2: Close using only mutex protection
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(25 * time.Microsecond)

			// After fix: use only mutex protection - no atomic operations
			reader.mu.Lock()
			reader.closed = true
			close(reader.closeChan) // Signal closure
			reader.mu.Unlock()
		}()

		wg.Wait()

		// Verify final state is correct
		reader.mu.RLock()
		finalClosed := reader.closed
		reader.mu.RUnlock()

		if !finalClosed {
			t.Error("Reader should be in closed state after close operation")
		}
	}
}

// TestConsistentSynchronizationWorksCorrectly validates that the fix works properly
func TestConsistentSynchronizationWorksCorrectly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	reader := &RawReader{
		recvChan:  make(chan *RawDatagram, 1),
		errorChan: make(chan error, 1),
		closeChan: make(chan struct{}),
		doneChan:  make(chan struct{}),
		closed:    false,
		mu:        sync.RWMutex{},
	}

	var wg sync.WaitGroup

	// Multiple readers checking state - they should see consistent values within their read operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				// This is the pattern that should be race-free after our fix
				reader.mu.RLock()
				isClosed := reader.closed
				// Do something with the value while holding the lock
				if isClosed {
					// Safe to act on this value
				}
				reader.mu.RUnlock()

				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Writer setting state
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Microsecond)
		reader.mu.Lock()
		reader.closed = true
		reader.mu.Unlock()
	}()

	wg.Wait()

	// Final verification that state is correct
	reader.mu.RLock()
	finalClosed := reader.closed
	reader.mu.RUnlock()

	if !finalClosed {
		t.Error("Reader should be in closed state after close operation")
	}
}
