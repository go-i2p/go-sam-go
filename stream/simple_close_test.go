package stream

import (
	"testing"
	"time"
)

// TestSimpleSessionClose tests basic session close functionality
func TestSimpleSessionClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a session
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_simple_close", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Create one listener
	listener, err := session.Listen()
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	// Give the listener time to start
	time.Sleep(100 * time.Millisecond)

	// Close the session - this should complete quickly
	start := time.Now()
	err = session.Close()
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Failed to close session: %v", err)
	}

	// Session close should complete within 1 second
	if duration > time.Second {
		t.Errorf("Session close took too long: %v", duration)
	}

	// Verify listener is closed
	_, err = listener.Accept()
	if err == nil {
		t.Error("Listener should be closed but Accept() succeeded")
	}
}
