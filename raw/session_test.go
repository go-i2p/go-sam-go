package raw

import (
	"testing"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

const testSAMAddr = "127.0.0.1:7656"

func setupTestSAM(t *testing.T) (*common.SAM, i2pkeys.I2PKeys) {
	t.Helper()

	// Add retry mechanism for resource exhaustion
	var sam *common.SAM
	var err error
	for attempts := 0; attempts < 3; attempts++ {
		sam, err = common.NewSAM(testSAMAddr)
		if err == nil {
			break
		}
		// Wait a bit before retrying in case of resource exhaustion
		time.Sleep(time.Duration(attempts+1) * 500 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("Failed to create SAM connection after retries: %v", err)
	}

	// Ensure SAM connection is closed when test completes
	t.Cleanup(func() {
		if sam != nil {
			sam.Close()
		}
	})

	keys, err := sam.NewKeys()
	if err != nil {
		sam.Close()
		t.Fatalf("Failed to generate keys: %v", err)
	}

	return sam, keys
}

func TestNewRawSession(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		options []string
		wantErr bool
	}{
		{
			name:    "basic session creation",
			id:      "test_raw_session",
			options: nil,
			wantErr: false,
		},
		{
			name:    "session with options",
			id:      "test_raw_with_opts",
			options: []string{"inbound.length=1", "outbound.length=1"},
			wantErr: false,
		},
		{
			name: "session with small tunnel config",
			id:   "test_raw_small",
			options: []string{
				"inbound.length=1",
				"outbound.length=1",
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

			session, err := NewRawSession(sam, tt.id, keys, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRawSession() error = %v, wantErr %v", err, tt.wantErr)
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

func TestRawSession_Close(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewRawSession(sam, "test_close", keys, nil)
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
}

func TestRawSession_Addr(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewRawSession(sam, "test_addr", keys, nil)
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

func TestRawSession_NewReader(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewRawSession(sam, "test_reader", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	reader := session.NewReader()
	if reader == nil {
		t.Error("NewReader() returned nil")
	}
	// Critical fix: Ensure reader is properly closed to prevent goroutine leaks
	defer reader.Close()

	if reader.session != session {
		t.Error("Reader session reference is incorrect")
	}

	// Verify channels are initialized
	if reader.recvChan == nil {
		t.Error("Reader recvChan is nil")
	}
	if reader.errorChan == nil {
		t.Error("Reader errorChan is nil")
	}
	if reader.closeChan == nil {
		t.Error("Reader closeChan is nil")
	}
	if reader.doneChan == nil {
		t.Error("Reader doneChan is nil")
	}
}

func TestRawSession_NewWriter(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewRawSession(sam, "test_writer", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	writer := session.NewWriter()
	if writer == nil {
		t.Error("NewWriter() returned nil")
	}

	if writer.session != session {
		t.Error("Writer session reference is incorrect")
	}

	if writer.timeout != 30 {
		t.Errorf("Writer timeout = %v, want 30", writer.timeout)
	}
}

func TestRawSession_PacketConn(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	session, err := NewRawSession(sam, "test_packetconn", keys, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	conn := session.PacketConn()
	if conn == nil {
		t.Error("PacketConn() returned nil")
	}
	// Critical fix: Ensure PacketConn is properly closed to prevent goroutine leaks
	defer conn.Close()

	rawConn, ok := conn.(*RawConn)
	if !ok {
		t.Error("PacketConn() did not return a RawConn")
	}

	if rawConn.session != session {
		t.Error("RawConn session reference is incorrect")
	}

	if rawConn.reader == nil {
		t.Error("RawConn reader is nil")
	}

	if rawConn.writer == nil {
		t.Error("RawConn writer is nil")
	}
}

func TestRawAddr_Network(t *testing.T) {
	addr := &RawAddr{}
	if addr.Network() != "i2p-raw" {
		t.Errorf("Network() = %v, want i2p-raw", addr.Network())
	}
}

func TestRawAddr_String(t *testing.T) {
	sam, keys := setupTestSAM(t)
	defer sam.Close()

	addr := &RawAddr{addr: keys.Addr()}
	expected := keys.Addr().Base32()

	if addr.String() != expected {
		t.Errorf("String() = %v, want %v", addr.String(), expected)
	}
}
