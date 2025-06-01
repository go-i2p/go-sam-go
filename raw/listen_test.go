package raw

import (
	"testing"
	"time"

	"github.com/go-i2p/go-sam-go/common"
)

func TestRawSession_Listen(t *testing.T) {
	tests := []struct {
		name         string
		setupSession func() *RawSession
		wantErr      bool
		errContains  string
	}{
		{
			name: "successful_listen",
			setupSession: func() *RawSession {
				// Create a mock SAM connection
				sam := &common.SAM{}
				baseSession := &common.BaseSession{}
				return &RawSession{
					BaseSession: baseSession,
					sam:         sam,
					options:     []string{},
					closed:      false,
				}
			},
			wantErr: false,
		},
		{
			name: "listen_on_closed_session",
			setupSession: func() *RawSession {
				sam := &common.SAM{}
				baseSession := &common.BaseSession{}
				return &RawSession{
					BaseSession: baseSession,
					sam:         sam,
					options:     []string{},
					closed:      true,
				}
			},
			wantErr:     true,
			errContains: "session is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := tt.setupSession()

			listener, err := session.Listen()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Listen() expected error but got none")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("Listen() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Listen() unexpected error = %v", err)
				return
			}

			if listener == nil {
				t.Error("Listen() returned nil listener")
				return
			}

			// Verify listener properties
			if listener.session != session {
				t.Error("Listener session reference incorrect")
			}

			if listener.reader == nil {
				t.Error("Listener reader not initialized")
			}

			if listener.acceptChan == nil {
				t.Error("Listener acceptChan not initialized")
			}

			if listener.errorChan == nil {
				t.Error("Listener errorChan not initialized")
			}

			if listener.closeChan == nil {
				t.Error("Listener closeChan not initialized")
			}

			// Clean up
			if listener != nil {
				_ = listener.Close()
			}
		})
	}
}

func TestRawListener_Properties(t *testing.T) {
	// Setup a basic session for testing
	sam := &common.SAM{}
	baseSession := &common.BaseSession{}
	session := &RawSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     []string{},
		closed:      false,
	}

	listener, err := session.Listen()
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	t.Run("channels_buffered_correctly", func(t *testing.T) {
		// Check that acceptChan has proper buffer size
		select {
		case listener.acceptChan <- &RawConn{}:
			// Should not block for first 10 items
		case <-time.After(100 * time.Millisecond):
			t.Error("acceptChan appears to be unbuffered or too small")
		}

		// Drain the channel
		select {
		case <-listener.acceptChan:
		default:
			t.Error("Failed to read from acceptChan")
		}
	})

	t.Run("initial_state", func(t *testing.T) {
		if listener.closed {
			t.Error("New listener should not be closed initially")
		}
	})
}

func TestRawListener_Close(t *testing.T) {
	sam := &common.SAM{}
	baseSession := &common.BaseSession{}
	session := &RawSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     []string{},
		closed:      false,
	}

	listener, err := session.Listen()
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	// Test closing
	err = listener.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Verify closed state
	listener.mu.RLock()
	closed := listener.closed
	listener.mu.RUnlock()

	if !closed {
		t.Error("Listener should be marked as closed after Close()")
	}

	// Test double close
	err = listener.Close()
	if err == nil {
		t.Error("Second Close() should return error")
	}
}

func TestRawListener_Concurrent_Access(t *testing.T) {
	sam := &common.SAM{}
	baseSession := &common.BaseSession{}
	session := &RawSession{
		BaseSession: baseSession,
		sam:         sam,
		options:     []string{},
		closed:      false,
	}

	listener, err := session.Listen()
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Test concurrent access to listener state
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			listener.mu.RLock()
			_ = listener.closed
			listener.mu.RUnlock()
		}
	}()

	go func() {
		defer func() { done <- true }()
		for i := 0; i < 100; i++ {
			listener.mu.RLock()
			_ = listener.session
			listener.mu.RUnlock()
		}
	}()

	// Wait for both goroutines
	<-done
	<-done
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(substr == "" || findString(s, substr))
}

func findString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
