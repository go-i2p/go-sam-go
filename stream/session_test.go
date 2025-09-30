package stream

import (
	"testing"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

const testSAMAddr = "127.0.0.1:7656"

func setupTestSAM(t *testing.T) (*common.SAM, i2pkeys.I2PKeys) {
	t.Helper()

	sam, err := common.NewSAM(testSAMAddr)
	if err != nil {
		t.Fatalf("Failed to create SAM connection: %v", err)
	}

	keys, err := sam.NewKeys()
	if err != nil {
		sam.Close()
		t.Fatalf("Failed to generate keys: %v", err)
	}

	return sam, keys
}

func TestNewStreamSession(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		options []string
		wantErr bool
	}{
		{
			name:    "basic session creation",
			id:      "test_stream_session",
			options: nil,
			wantErr: false,
		},
		{
			name:    "session with options",
			id:      "test_stream_with_opts",
			options: []string{"inbound.length=1", "outbound.length=1"},
			wantErr: false,
		},
		{
			name: "session with small tunnel config",
			id:   "test_stream_small",
			options: []string{
				"inbound.length=0",
				"outbound.length=0",
				"inbound.lengthVariance=0",
				"outbound.lengthVariance=0",
				"inbound.quantity=1",
				"outbound.quantity=1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sam, keys := setupTestSAM(t)
			defer sam.Close()

			session, err := NewStreamSession(sam, tt.id, keys, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStreamSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// Verify session properties
				if session.ID() != tt.id {
					t.Errorf("Session ID = %v, want %v", session.ID(), tt.id)
				}

				if session.Keys().Addr().Base32() != keys.Addr().Base32() {
					t.Error("Session keys don't match provided keys")
				}

				addr := session.Addr()
				if addr.Base32() == "" {
					t.Error("Session address is empty")
				}

				// Clean up
				if err := session.Close(); err != nil {
					t.Errorf("Failed to close session: %v", err)
				}
			}
		})
	}
}

func TestStreamSession_Close(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_close", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Close the session
	err = session.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Closing again should not error
	err = session.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}

	// Operations on closed session should fail
	_, err = session.Listen()
	if err == nil {
		t.Error("Listen() on closed session should fail")
	}
}

func TestStreamSession_Addr(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_addr", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	addr := session.Addr()
	expectedAddr := keys.Addr()

	if addr.Base32() != expectedAddr.Base32() {
		t.Errorf("Addr() = %v, want %v", addr.Base32(), expectedAddr.Base32())
	}
}

func TestStreamSession_ConcurrentOperations(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewStreamSession(sam, "test_concurrent", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Test concurrent dialer creation
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			dialer := session.NewDialer()
			if dialer == nil {
				t.Error("NewDialer returned nil")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
