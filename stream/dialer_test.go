package stream

import (
	"context"
	"testing"
	"time"
)

func TestStreamSession_Dial(t *testing.T) {

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_dial", keys, []string{
		"inbound.length=1", "outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Test dialing to a known I2P destination
	// This test might fail if the destination is not reachable
	// but it tests the basic dial functionality
	_, err = session.Dial("idk.i2p")
	// We don't fail the test if dial fails since it depends on network conditions
	// but we log it for debugging
	if err != nil {
		t.Logf("Dial to idk.i2p failed (expected in some network conditions): %v", err)
	}
}

func TestStreamSession_DialI2P(t *testing.T) {

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_dial_i2p", keys, []string{
		"inbound.length=1", "outbound.length=1",
	})
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Try to lookup a destination first
	addr, err := sam.Lookup("zzz.i2p")
	if err != nil {
		t.Skipf("Failed to lookup destination: %v", err)
	}

	// Test dialing to the looked up address
	_, err = session.DialI2P(addr)
	if err != nil {
		t.Logf("DialI2P failed (expected in some network conditions): %v", err)
	}
}

func TestStreamSession_DialContext(t *testing.T) {

	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_dial_context", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	t.Run("dial with context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := session.DialContext(ctx, "nonexistent.i2p")
		if err == nil {
			t.Log("Dial succeeded unexpectedly")
		} else {
			t.Logf("Dial failed as expected: %v", err)
		}
	})

	t.Run("dial with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := session.DialContext(ctx, "test.i2p")
		if err == nil {
			t.Error("Expected dial to fail with cancelled context")
		}
	})
}
