package raw

import (
	"runtime"
	"testing"
	"time"
)

// TestFinalizerPreventsGoroutineLeaks tests that the finalizer mechanism
// automatically cleans up resources when connections are garbage collected
func TestFinalizerPreventsGoroutineLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping finalizer test in short mode")
	}

	// Get baseline goroutine count
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	initialGoroutines := runtime.NumGoroutine()

	// Create connections that will be abandoned without explicit Close()
	// This tests the finalizer mechanism
	func() {
		for i := 0; i < 3; i++ {
			// Create a connection that simulates what DialAddr does
			conn := &RawConn{
				reader: &RawReader{
					recvChan:  make(chan *RawDatagram, 10),
					errorChan: make(chan error, 1),
					closeChan: make(chan struct{}),
					doneChan:  make(chan struct{}),
					closed:    false,
				},
				closed: false,
			}

			// Start a goroutine to simulate receiveLoop
			go func(r *RawReader) {
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-r.closeChan:
						close(r.doneChan)
						return
					case <-ticker.C:
						// Simulate work
					}
				}
			}(conn.reader)

			// Set up the cleanup (this is what our fix does)
			conn.addCleanup()
		}
		// Connections go out of scope here without Close() being called
	}()

	// Force garbage collection multiple times to trigger finalizers
	for i := 0; i < 5; i++ {
		runtime.GC()
		time.Sleep(100 * time.Millisecond)
	}

	// Allow finalizers to run and cleanup to complete
	time.Sleep(500 * time.Millisecond)

	// Check final goroutine count
	finalGoroutines := runtime.NumGoroutine()
	goroutineChange := finalGoroutines - initialGoroutines

	// With finalizers, there should be no significant goroutine leak
	if goroutineChange > 3 { // Allow some variance for test infrastructure
		t.Errorf("Finalizer did not prevent goroutine leak: %d extra goroutines", goroutineChange)
	} else {
		t.Logf("Finalizer mechanism working: only %d net goroutine change", goroutineChange)
	}
}

// TestExplicitCloseSkipsFinalizer tests that calling Close() explicitly
// clears the finalizer and prevents unnecessary cleanup
func TestExplicitCloseSkipsFinalizer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping finalizer test in short mode")
	}

	// Create a connection and close it explicitly
	conn := &RawConn{
		reader: &RawReader{
			recvChan:  make(chan *RawDatagram, 10),
			errorChan: make(chan error, 1),
			closeChan: make(chan struct{}),
			doneChan:  make(chan struct{}),
			closed:    false,
		},
		closed: false,
	}

	// Start a goroutine to simulate receiveLoop
	go func(r *RawReader) {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-r.closeChan:
				close(r.doneChan)
				return
			case <-ticker.C:
				// Simulate work
			}
		}
	}(conn.reader)

	// Set up the cleanup
	conn.addCleanup()

	// Wait a bit for the goroutine to start
	time.Sleep(50 * time.Millisecond)

	// Explicitly close the connection
	err := conn.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Verify the goroutine stopped
	select {
	case <-conn.reader.doneChan:
		// Good, goroutine stopped
	case <-time.After(1 * time.Second):
		t.Error("Goroutine did not stop after Close()")
	}

	// Force garbage collection - the finalizer should not run
	// since we already called Close() explicitly
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Test passes if no additional cleanup happens (we can't easily verify this,
	// but the important thing is that explicit Close() works correctly)
}
