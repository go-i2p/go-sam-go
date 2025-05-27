package stream

import (
	"testing"
	"time"
)

func TestStreamSession_Listen(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_listen", keys, []string{
		"inbound.length=0", "outbound.length=0",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	listener, err := session.Listen()
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Verify listener properties
	if listener.Addr().String() != session.Addr().String() {
		t.Error("Listener address doesn't match session address")
	}
}

func TestStreamSession_NewDialer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_dialer", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	dialer := session.NewDialer()
	if dialer == nil {
		t.Fatal("NewDialer returned nil")
	}

	// Test setting timeout
	newTimeout := 45 * time.Second
	dialer.SetTimeout(newTimeout)
	if dialer.timeout != newTimeout {
		t.Errorf("SetTimeout() timeout = %v, want %v", dialer.timeout, newTimeout)
	}
}
