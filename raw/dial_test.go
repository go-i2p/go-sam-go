package raw

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// generateUniqueSessionID creates a unique session ID to prevent conflicts during concurrent test execution.
// This ensures test isolation when multiple tests run simultaneously (e.g., during race detection).
func generateUniqueSessionID(testName string) string {
	// Use timestamp (nanoseconds) and random number to ensure uniqueness across concurrent executions
	timestamp := time.Now().UnixNano()
	random := rand.Intn(99999)
	return fmt.Sprintf("%s_%d_%05d", testName, timestamp, random)
}

func setupTestSession(t *testing.T, testName string) *RawSession {
	t.Helper()

	// Skip actual I2P connection for unit tests

	sam, err := common.NewSAM(testSAMAddr)
	if err != nil {
		t.Fatalf("Failed to create SAM connection: %v", err)
	}

	keys, err := sam.NewKeys()
	if err != nil {
		sam.Close()
		t.Fatalf("Failed to generate keys: %v", err)
	}

	sessionID := generateUniqueSessionID(testName)
	session, err := NewRawSession(sam, sessionID, keys, nil)
	if err != nil {
		sam.Close()
		t.Fatalf("Failed to create session: %v", err)
	}

	return session
}

// Update the test to use proper session setup
func TestRawSession_Dial(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid_b32_destination",
			destination: "example.b32.i2p",
			wantErr:     false,
		},
		{
			name:        "empty_destination",
			destination: "",
			wantErr:     true,
			errContains: "destination",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Short() {
				t.Skip("Skipping integration test in short mode")
			}

			session := setupTestSession(t, tt.name)
			defer session.Close()

			conn, err := session.Dial(tt.destination)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Dial() expected error but got none")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("Dial() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Dial() unexpected error = %v", err)
				return
			}

			if conn == nil {
				t.Error("Dial() returned nil connection")
				return
			}

			// Clean up
			if conn != nil {
				_ = conn.Close()
			}
		})
	}
}

func TestRawSession_DialTimeout(t *testing.T) {
	tests := []struct {
		name         string
		destination  string
		timeout      time.Duration
		setupSession func() *RawSession
		wantErr      bool
		errContains  string
	}{
		{
			name:        "valid_dial_with_timeout",
			destination: "example.b32.i2p",
			timeout:     5 * time.Second,
			setupSession: func() *RawSession {
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
			name:        "zero_timeout",
			destination: "example.b32.i2p",
			timeout:     0,
			setupSession: func() *RawSession {
				sam := &common.SAM{}
				baseSession := &common.BaseSession{}
				return &RawSession{
					BaseSession: baseSession,
					sam:         sam,
					options:     []string{},
					closed:      false,
				}
			},
			wantErr: false, // Zero timeout should still work, just no timeout
		},
		{
			name:        "negative_timeout",
			destination: "example.b32.i2p",
			timeout:     -1 * time.Second,
			setupSession: func() *RawSession {
				sam := &common.SAM{}
				baseSession := &common.BaseSession{}
				return &RawSession{
					BaseSession: baseSession,
					sam:         sam,
					options:     []string{},
					closed:      false,
				}
			},
			wantErr: false, // Implementation should handle negative timeout gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := tt.setupSession()

			conn, err := session.DialTimeout(tt.destination, tt.timeout)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DialTimeout() expected error but got none")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("DialTimeout() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("DialTimeout() unexpected error = %v", err)
				return
			}

			if conn == nil {
				t.Error("DialTimeout() returned nil connection")
				return
			}

			// Verify conn implements net.PacketConn
			if _, ok := conn.(net.PacketConn); !ok {
				t.Error("DialTimeout() returned connection that doesn't implement net.PacketConn")
			}

			// Clean up
			if conn != nil {
				_ = conn.Close()
			}
		})
	}
}

func TestRawSession_DialContext(t *testing.T) {
	tests := []struct {
		name         string
		destination  string
		setupContext func() context.Context
		setupSession func() *RawSession
		wantErr      bool
		errContains  string
	}{
		{
			name:        "valid_dial_with_context",
			destination: "example.b32.i2p",
			setupContext: func() context.Context {
				return context.Background()
			},
			setupSession: func() *RawSession {
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
			name:        "cancelled_context",
			destination: "example.b32.i2p",
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx
			},
			setupSession: func() *RawSession {
				sam := &common.SAM{}
				baseSession := &common.BaseSession{}
				return &RawSession{
					BaseSession: baseSession,
					sam:         sam,
					options:     []string{},
					closed:      false,
				}
			},
			wantErr:     true,
			errContains: "context",
		},
		{
			name:        "context_with_timeout",
			destination: "example.b32.i2p",
			setupContext: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
				return ctx
			},
			setupSession: func() *RawSession {
				sam := &common.SAM{}
				baseSession := &common.BaseSession{}
				return &RawSession{
					BaseSession: baseSession,
					sam:         sam,
					options:     []string{},
					closed:      false,
				}
			},
			wantErr: false, // Should succeed if dial completes quickly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := tt.setupSession()
			ctx := tt.setupContext()

			conn, err := session.DialContext(ctx, tt.destination)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DialContext() expected error but got none")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("DialContext() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("DialContext() unexpected error = %v", err)
				return
			}

			if conn == nil {
				t.Error("DialContext() returned nil connection")
				return
			}

			// Verify conn implements net.PacketConn
			if _, ok := conn.(net.PacketConn); !ok {
				t.Error("DialContext() returned connection that doesn't implement net.PacketConn")
			}

			// Clean up
			if conn != nil {
				_ = conn.Close()
			}
		})
	}
}

func TestRawSession_DialI2P(t *testing.T) {
	// Create a test I2P address
	testAddr := createTestI2PAddr()

	tests := []struct {
		name         string
		addr         i2pkeys.I2PAddr
		setupSession func() *RawSession
		wantErr      bool
		errContains  string
	}{
		{
			name: "valid_i2p_address",
			addr: testAddr,
			setupSession: func() *RawSession {
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
			name: "dial_i2p_on_closed_session",
			addr: testAddr,
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
			errContains: "closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := tt.setupSession()

			conn, err := session.DialI2P(tt.addr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DialI2P() expected error but got none")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("DialI2P() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("DialI2P() unexpected error = %v", err)
				return
			}

			if conn == nil {
				t.Error("DialI2P() returned nil connection")
				return
			}

			// Verify conn implements net.PacketConn
			if _, ok := conn.(net.PacketConn); !ok {
				t.Error("DialI2P() returned connection that doesn't implement net.PacketConn")
			}

			// Clean up
			if conn != nil {
				_ = conn.Close()
			}
		})
	}
}

func TestRawSession_DialI2PTimeout(t *testing.T) {
	testAddr := createTestI2PAddr()

	tests := []struct {
		name         string
		addr         i2pkeys.I2PAddr
		timeout      time.Duration
		setupSession func() *RawSession
		wantErr      bool
		errContains  string
	}{
		{
			name:    "valid_i2p_dial_with_timeout",
			addr:    testAddr,
			timeout: 5 * time.Second,
			setupSession: func() *RawSession {
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
			name:    "i2p_dial_zero_timeout",
			addr:    testAddr,
			timeout: 0,
			setupSession: func() *RawSession {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := tt.setupSession()

			conn, err := session.DialI2PTimeout(tt.addr, tt.timeout)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DialI2PTimeout() expected error but got none")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("DialI2PTimeout() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("DialI2PTimeout() unexpected error = %v", err)
				return
			}

			if conn == nil {
				t.Error("DialI2PTimeout() returned nil connection")
				return
			}

			// Verify conn implements net.PacketConn
			if _, ok := conn.(net.PacketConn); !ok {
				t.Error("DialI2PTimeout() returned connection that doesn't implement net.PacketConn")
			}

			// Clean up
			if conn != nil {
				_ = conn.Close()
			}
		})
	}
}

func TestRawSession_DialI2PContext(t *testing.T) {
	testAddr := createTestI2PAddr()

	tests := []struct {
		name         string
		addr         i2pkeys.I2PAddr
		setupContext func() context.Context
		setupSession func() *RawSession
		wantErr      bool
		errContains  string
	}{
		{
			name: "valid_i2p_dial_with_context",
			addr: testAddr,
			setupContext: func() context.Context {
				return context.Background()
			},
			setupSession: func() *RawSession {
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
			name: "i2p_dial_cancelled_context",
			addr: testAddr,
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupSession: func() *RawSession {
				sam := &common.SAM{}
				baseSession := &common.BaseSession{}
				return &RawSession{
					BaseSession: baseSession,
					sam:         sam,
					options:     []string{},
					closed:      false,
				}
			},
			wantErr:     true,
			errContains: "context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := tt.setupSession()
			ctx := tt.setupContext()

			conn, err := session.DialI2PContext(ctx, tt.addr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DialI2PContext() expected error but got none")
					return
				}
				if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("DialI2PContext() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("DialI2PContext() unexpected error = %v", err)
				return
			}

			if conn == nil {
				t.Error("DialI2PContext() returned nil connection")
				return
			}

			// Verify conn implements net.PacketConn
			if _, ok := conn.(net.PacketConn); !ok {
				t.Error("DialI2PContext() returned connection that doesn't implement net.PacketConn")
			}

			// Clean up
			if conn != nil {
				_ = conn.Close()
			}
		})
	}
}

// Helper function to create a test I2P address
func createTestI2PAddr() i2pkeys.I2PAddr {
	// Create a minimal test address - in real implementation this would be a proper I2P destination
	dest, _ := i2pkeys.NewDestination()
	return dest.Addr()
}
