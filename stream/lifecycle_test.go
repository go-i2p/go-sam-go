package stream

import (
	"runtime"
	"testing"
	"time"
)

// TestSessionListenerLifecycle tests that listeners are properly cleaned up when session closes
func TestSessionListenerLifecycle(t *testing.T) {

	// Create a session
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_lifecycle", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Get initial goroutine count
	initialGoroutines := runtime.NumGoroutine()

	// Create multiple listeners
	var listeners []*StreamListener
	for i := 0; i < 3; i++ {
		listener, err := session.Listen()
		if err != nil {
			t.Fatalf("Failed to create listener %d: %v", i, err)
		}
		listeners = append(listeners, listener)
	}

	// Wait for goroutines to start
	time.Sleep(100 * time.Millisecond)

	// Verify goroutines increased
	afterCreateGoroutines := runtime.NumGoroutine()
	if afterCreateGoroutines <= initialGoroutines {
		t.Errorf("Expected more goroutines after creating listeners, got %d, was %d", afterCreateGoroutines, initialGoroutines)
	}

	// Close the session - this should clean up all listeners
	err = session.Close()
	if err != nil {
		t.Errorf("Failed to close session: %v", err)
	}

	// Wait for cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify goroutines were cleaned up
	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > afterCreateGoroutines {
		t.Errorf("Expected fewer or same goroutines after closing session, got %d, was %d", finalGoroutines, afterCreateGoroutines)
	}

	// Verify listeners are actually closed
	for i, listener := range listeners {
		// Try to use the listener - should fail
		_, err := listener.Accept()
		if err == nil {
			t.Errorf("Listener %d should be closed but Accept() succeeded", i)
		}
	}
}

// TestExplicitListenerClose tests that explicitly closing listeners works correctly
func TestExplicitListenerClose(t *testing.T) {

	// Create a session
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_explicit_close", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Create a listener
	listener, err := session.Listen()
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	// Close the listener explicitly
	err = listener.Close()
	if err != nil {
		t.Errorf("Failed to close listener: %v", err)
	}

	// Verify listener is closed
	_, err = listener.Accept()
	if err == nil {
		t.Error("Listener should be closed but Accept() succeeded")
	}
}
