package common

import (
	"strings"
	"testing"

	"github.com/go-i2p/i2pkeys"
)

func TestNewGenericSession(t *testing.T) {
	// Create SAM connection
	sam, err := NewSAM("127.0.0.1:7656")
	if err != nil {
		t.Skipf("Failed to connect to SAM bridge: %v", err)
	}
	defer sam.Close()

	// Generate keys for testing
	_, err = sam.NewKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	tests := []struct {
		name    string
		style   string
		id      string
		extras  []string
		wantErr bool
	}{
		{
			name:    "stream session",
			style:   SESSION_STYLE_STREAM,
			id:      "test-stream-session",
			extras:  []string{},
			wantErr: false,
		},
		{
			name:    "datagram session",
			style:   SESSION_STYLE_DATAGRAM,
			id:      "test-datagram-session",
			extras:  []string{},
			wantErr: false,
		},
		{
			name:    "raw session",
			style:   SESSION_STYLE_RAW,
			id:      "test-raw-session",
			extras:  []string{},
			wantErr: false,
		},
		{
			name:    "invalid style",
			style:   "INVALID",
			id:      "test-invalid-session",
			extras:  []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new SAM connection for each test
			testSam, err := NewSAM("127.0.0.1:7656")
			if err != nil {
				t.Skipf("Failed to connect to SAM bridge: %v", err)
			}
			defer testSam.Close()

			// Generate unique keys for each test
			testKeys, err := testSam.NewKeys()
			if err != nil {
				t.Fatalf("Failed to generate keys: %v", err)
			}

			session, err := testSam.NewGenericSession(tt.style, tt.id, testKeys, tt.extras)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenericSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if session == nil {
					t.Error("NewGenericSession() returned nil session")
					return
				}

				// Verify session properties
				if session.ID() != tt.id {
					t.Errorf("Session ID = %v, want %v", session.ID(), tt.id)
				}

				if session.Keys().String() != testKeys.String() {
					t.Error("Session keys don't match expected keys")
				}

				// Clean up session
				session.Close()
			}
		})
	}
}

func TestNewGenericSessionWithSignature(t *testing.T) {
	// Create SAM connection
	sam, err := NewSAM("127.0.0.1:7656")
	if err != nil {
		t.Skipf("Failed to connect to SAM bridge: %v", err)
	}
	defer sam.Close()

	tests := []struct {
		name    string
		style   string
		id      string
		sigType string
		extras  []string
		wantErr bool
	}{
		{
			name:    "ed25519 signature",
			style:   SESSION_STYLE_STREAM,
			id:      "test-ed25519-session",
			sigType: SIG_EdDSA_SHA512_Ed25519,
			extras:  []string{},
			wantErr: false,
		},
		{
			name:    "dsa signature",
			style:   SESSION_STYLE_STREAM,
			id:      "test-dsa-session",
			sigType: SIG_DSA_SHA1,
			extras:  []string{},
			wantErr: false,
		},
		{
			name:    "ecdsa p256 signature",
			style:   SESSION_STYLE_STREAM,
			id:      "test-ecdsa-p256-session",
			sigType: SIG_ECDSA_SHA256_P256,
			extras:  []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new SAM connection for each test
			testSam, err := NewSAM("127.0.0.1:7656")
			if err != nil {
				t.Skipf("Failed to connect to SAM bridge: %v", err)
			}
			defer testSam.Close()

			// Generate keys with specific signature type
			var testKeys i2pkeys.I2PKeys
			if strings.Contains(tt.sigType, "EdDSA") {
				testKeys, err = testSam.NewKeys("EdDSA_SHA512_Ed25519")
			} else if strings.Contains(tt.sigType, "DSA") {
				testKeys, err = testSam.NewKeys("DSA_SHA1")
			} else if strings.Contains(tt.sigType, "ECDSA_SHA256_P256") {
				testKeys, err = testSam.NewKeys("ECDSA_SHA256_P256")
			} else {
				testKeys, err = testSam.NewKeys()
			}
			if err != nil {
				t.Fatalf("Failed to generate keys: %v", err)
			}

			session, err := testSam.NewGenericSessionWithSignature(tt.style, tt.id, testKeys, tt.sigType, tt.extras)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenericSessionWithSignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if session == nil {
					t.Error("NewGenericSessionWithSignature() returned nil session")
					return
				}

				// Verify session properties
				if session.ID() != tt.id {
					t.Errorf("Session ID = %v, want %v", session.ID(), tt.id)
				}

				// Clean up session
				session.Close()
			}
		})
	}
}

func TestNewGenericSessionWithSignatureAndPorts(t *testing.T) {
	// Create SAM connection
	sam, err := NewSAM("127.0.0.1:7656")
	if err != nil {
		t.Skipf("Failed to connect to SAM bridge: %v", err)
	}
	defer sam.Close()

	tests := getSessionWithPortsTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSam, testKeys := setupSessionTest(t)
			defer testSam.Close()

			session, err := testSam.NewGenericSessionWithSignatureAndPorts(tt.style, tt.id, tt.from, tt.to, testKeys, tt.sigType, tt.extras)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGenericSessionWithSignatureAndPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				validateSessionWithPorts(t, session, tt, testKeys)
				session.Close()
			}
		})
	}
}

// getSessionWithPortsTestCases returns test cases for session creation with ports and signature.
func getSessionWithPortsTestCases() []struct {
	name    string
	style   string
	id      string
	from    string
	to      string
	sigType string
	extras  []string
	wantErr bool
} {
	return []struct {
		name    string
		style   string
		id      string
		from    string
		to      string
		sigType string
		extras  []string
		wantErr bool
	}{
		{
			name:    "with ports",
			style:   SESSION_STYLE_STREAM,
			id:      "test-ports-session",
			from:    "8080",
			to:      "9090",
			sigType: SIG_EdDSA_SHA512_Ed25519,
			extras:  []string{},
			wantErr: false,
		},
		{
			name:    "default ports",
			style:   SESSION_STYLE_STREAM,
			id:      "test-default-ports-session",
			from:    "0",
			to:      "0",
			sigType: SIG_EdDSA_SHA512_Ed25519,
			extras:  []string{},
			wantErr: false,
		},
		{
			name:    "with extras",
			style:   SESSION_STYLE_STREAM,
			id:      "test-extras-session",
			from:    "0",
			to:      "0",
			sigType: SIG_EdDSA_SHA512_Ed25519,
			extras:  []string{"inbound.length=2", "outbound.length=2"},
			wantErr: false,
		},
	}
}

// setupSessionTest creates a new SAM connection and generates test keys.
func setupSessionTest(t *testing.T) (*SAM, i2pkeys.I2PKeys) {
	testSam, err := NewSAM("127.0.0.1:7656")
	if err != nil {
		t.Skipf("Failed to connect to SAM bridge: %v", err)
	}

	testKeys, err := testSam.NewKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	return testSam, testKeys
}

// validateSessionWithPorts verifies that a session was created correctly with the expected properties.
func validateSessionWithPorts(t *testing.T, session Session, testCase struct {
	name    string
	style   string
	id      string
	from    string
	to      string
	sigType string
	extras  []string
	wantErr bool
}, testKeys i2pkeys.I2PKeys,
) {
	if session == nil {
		t.Error("NewGenericSessionWithSignatureAndPorts() returned nil session")
		return
	}

	validateSessionID(t, session, testCase.id)
	validateSessionKeys(t, session, testKeys)
	validateSessionPorts(t, session, testCase.from, testCase.to)
}

// validateSessionID checks if the session ID matches the expected value.
func validateSessionID(t *testing.T, session Session, expectedID string) {
	if session.ID() != expectedID {
		t.Errorf("Session ID = %v, want %v", session.ID(), expectedID)
	}
}

// validateSessionKeys checks if the session keys match the expected keys.
func validateSessionKeys(t *testing.T, session Session, expectedKeys i2pkeys.I2PKeys) {
	if session.Keys().String() != expectedKeys.String() {
		t.Error("Session keys don't match expected keys")
	}
}

// validateSessionPorts verifies that the session ports are set correctly.
func validateSessionPorts(t *testing.T, session Session, expectedFrom, expectedTo string) {
	baseSession, ok := session.(*BaseSession)
	if !ok {
		return
	}

	if expectedFrom != "0" && baseSession.From() != expectedFrom {
		t.Errorf("Session FromPort = %v, want %v", baseSession.From(), expectedFrom)
	}
	if expectedTo != "0" && baseSession.To() != expectedTo {
		t.Errorf("Session ToPort = %v, want %v", baseSession.To(), expectedTo)
	}
}

func TestSessionCreationErrors(t *testing.T) {
	// Create SAM connection
	sam, err := NewSAM("127.0.0.1:7656")
	if err != nil {
		t.Skipf("Failed to connect to SAM bridge: %v", err)
	}
	defer sam.Close()

	// Generate keys for testing
	keys, err := sam.NewKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys: %v", err)
	}

	t.Run("duplicate session ID", func(t *testing.T) {
		// Create first session
		testSam1, err := NewSAM("127.0.0.1:7656")
		if err != nil {
			t.Skipf("Failed to connect to SAM bridge: %v", err)
		}
		defer testSam1.Close()

		session1, err := testSam1.NewGenericSession(SESSION_STYLE_STREAM, "duplicate-id", keys, []string{})
		if err != nil {
			t.Fatalf("Failed to create first session: %v", err)
		}
		defer session1.Close()

		// Try to create second session with same ID
		testSam2, err := NewSAM("127.0.0.1:7656")
		if err != nil {
			t.Skipf("Failed to connect to SAM bridge: %v", err)
		}
		defer testSam2.Close()

		_, err = testSam2.NewGenericSession(SESSION_STYLE_STREAM, "duplicate-id", keys, []string{})
		if err == nil {
			t.Error("Expected error for duplicate session ID")
		}

		if !strings.Contains(err.Error(), "Duplicate") {
			t.Errorf("Expected duplicate error, got: %v", err)
		}
	})

	t.Run("invalid keys", func(t *testing.T) {
		testSam, err := NewSAM("127.0.0.1:7656")
		if err != nil {
			t.Skipf("Failed to connect to SAM bridge: %v", err)
		}
		defer testSam.Close()

		// Create invalid keys
		invalidKeys := i2pkeys.NewKeys(i2pkeys.I2PAddr("invalid"), "invalid")

		_, err = testSam.NewGenericSession(SESSION_STYLE_STREAM, "invalid-keys-session", invalidKeys, []string{})
		if err == nil {
			t.Error("Expected error for invalid keys")
		}
	})
}
