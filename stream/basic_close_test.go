package stream

import (
	"testing"
	"time"
)

// TestBasicSessionClose tests basic session close without listeners
func TestBasicSessionClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a session
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_basic_close", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Close the session without creating any listeners
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
}
