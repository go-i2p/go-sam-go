package datagram

import (
	"testing"
)

func TestDatagramSession_Listen(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewDatagramSession(sam, "test_listen", keys, []string{
		"inbound.length=1", "outbound.length=1",
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

	// Verify listener is not nil and has expected fields
	if listener.session != session {
		t.Error("Listener session doesn't match created session")
	}

	if listener.reader == nil {
		t.Error("Listener reader is nil")
	}

	if listener.acceptChan == nil {
		t.Error("Listener acceptChan is nil")
	}

	if listener.errorChan == nil {
		t.Error("Listener errorChan is nil")
	}

	if listener.closeChan == nil {
		t.Error("Listener closeChan is nil")
	}
}

func TestDatagramSession_Listen_ClosedSession(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewDatagramSession(sam, "test_listen_closed", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Close the session first
	session.Close()

	// Try to create listener on closed session
	listener, err := session.Listen()
	if err == nil {
		if listener != nil {
			listener.Close()
		}
		t.Fatal("Expected error when creating listener on closed session")
	}

	if listener != nil {
		t.Error("Expected nil listener when session is closed")
	}
}
