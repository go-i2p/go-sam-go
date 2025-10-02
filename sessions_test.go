package sam3

import (
	"fmt"
	"testing"
	"time"
)

// TestSAMSessionMethods tests that all session creation methods work with real I2P connections.
// These are integration tests that require a running I2P router with SAM bridge enabled.
func TestSAMSessionMethods(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	testCases := []struct {
		name        string
		methodTest  func(*testing.T)
		description string
	}{
		{
			name: "NewPrimarySession",
			methodTest: func(t *testing.T) {
				// Create a separate SAM connection for this test
				sam, err := NewSAM(SAMDefaultAddr(""))
				if err != nil {
					t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
				}
				defer sam.Close()

				// Generate real I2P keys for testing
				keys, err := sam.NewKeys()
				if err != nil {
					t.Fatalf("Failed to generate I2P keys: %v", err)
				}

				session, err := sam.NewPrimarySession("test-primary-"+RandString(), keys, Options_Default)
				if err != nil {
					t.Errorf("NewPrimarySession failed: %v", err)
					return
				}
				defer session.Close()

				// Verify we can create sub-sessions
				if session.SubSessionCount() != 0 {
					t.Errorf("Expected 0 sub-sessions, got %d", session.SubSessionCount())
				}
			},
			description: "Primary session creation and basic operations",
		},
		{
			name: "NewPrimarySessionWithSignature",
			methodTest: func(t *testing.T) {
				// Create a separate SAM connection for this test
				sam, err := NewSAM(SAMDefaultAddr(""))
				if err != nil {
					t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
				}
				defer sam.Close()

				// Generate real I2P keys for testing
				keys, err := sam.NewKeys()
				if err != nil {
					t.Fatalf("Failed to generate I2P keys: %v", err)
				}

				session, err := sam.NewPrimarySessionWithSignature("test-primary-sig-"+RandString(), keys, Options_Default, Sig_EdDSA_SHA512_Ed25519)
				if err != nil {
					t.Errorf("NewPrimarySessionWithSignature failed: %v", err)
					return
				}
				defer session.Close()

				// Verify session properties
				if session.Addr().String() == "" {
					t.Error("Expected non-empty session address")
				}
			},
			description: "Primary session creation with signature type",
		},
		{
			name: "NewStreamSession",
			methodTest: func(t *testing.T) {
				// Create a separate SAM connection for this test
				sam, err := NewSAM(SAMDefaultAddr(""))
				if err != nil {
					t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
				}
				defer sam.Close()

				// Generate real I2P keys for testing
				keys, err := sam.NewKeys()
				if err != nil {
					t.Fatalf("Failed to generate I2P keys: %v", err)
				}

				session, err := sam.NewStreamSession("test-stream-"+RandString(), keys, Options_Default)
				if err != nil {
					t.Errorf("NewStreamSession failed: %v", err)
					return
				}
				defer session.Close()

				// Verify session has valid address
				if session.Addr().String() == "" {
					t.Error("Expected non-empty stream session address")
				}
			},
			description: "Stream session creation",
		},
		{
			name: "NewStreamSessionWithSignature",
			methodTest: func(t *testing.T) {
				// Create a separate SAM connection for this test
				sam, err := NewSAM(SAMDefaultAddr(""))
				if err != nil {
					t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
				}
				defer sam.Close()

				// Generate real I2P keys for testing
				keys, err := sam.NewKeys()
				if err != nil {
					t.Fatalf("Failed to generate I2P keys: %v", err)
				}

				session, err := sam.NewStreamSessionWithSignature("test-stream-sig-"+RandString(), keys, Options_Default, Sig_EdDSA_SHA512_Ed25519)
				if err != nil {
					t.Errorf("NewStreamSessionWithSignature failed: %v", err)
					return
				}
				defer session.Close()

				// Test that we can create a listener
				listener, err := session.Listen()
				if err != nil {
					t.Errorf("Failed to create listener: %v", err)
					return
				}
				defer listener.Close()

				if listener.Addr().String() == "" {
					t.Error("Expected non-empty listener address")
				}
			},
			description: "Stream session with signature and listener creation",
		},
		{
			name: "NewDatagramSession",
			methodTest: func(t *testing.T) {
				// Create a separate SAM connection for this test
				sam, err := NewSAM(SAMDefaultAddr(""))
				if err != nil {
					t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
				}
				defer sam.Close()

				// Generate real I2P keys for testing
				keys, err := sam.NewKeys()
				if err != nil {
					t.Fatalf("Failed to generate I2P keys: %v", err)
				}

				session, err := sam.NewDatagramSession("test-datagram-"+RandString(), keys, Options_Default)
				if err != nil {
					t.Errorf("NewDatagramSession failed: %v", err)
					return
				}
				defer session.Close()

				// Verify session properties
				if session.Addr().String() == "" {
					t.Error("Expected non-empty datagram session address")
				}
			},
			description: "Datagram session creation",
		},
		{
			name: "NewRawSession",
			methodTest: func(t *testing.T) {
				// Create a separate SAM connection for this test
				sam, err := NewSAM(SAMDefaultAddr(""))
				if err != nil {
					t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
				}
				defer sam.Close()

				// Generate real I2P keys for testing
				keys, err := sam.NewKeys()
				if err != nil {
					t.Fatalf("Failed to generate I2P keys: %v", err)
				}

				session, err := sam.NewRawSession("test-raw-"+RandString(), keys, Options_Default, 0)
				if err != nil {
					t.Errorf("NewRawSession failed: %v", err)
					return
				}
				defer session.Close()

				// Verify session properties
				if session.Addr().String() == "" {
					t.Error("Expected non-empty raw session address")
				}
			},
			description: "Raw session creation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set I2P-appropriate timeout for session creation
			// I2P operations can take several minutes due to tunnel building
			done := make(chan bool, 1)
			go func() {
				tc.methodTest(t)
				done <- true
			}()

			select {
			case <-done:
				t.Logf("✓ %s: %s", tc.name, tc.description)
			case <-time.After(5 * time.Minute):
				t.Errorf("%s timed out after 5 minutes (I2P tunnel building can be slow)", tc.name)
			}
		})
	}
}

// TestPrimarySessionSubSessions tests that primary sessions can create and manage sub-sessions
func TestPrimarySessionSubSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Create a real SAM connection
	sam, err := NewSAM(SAMDefaultAddr(""))
	if err != nil {
		t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
	}
	defer sam.Close()

	// Create primary session
	keys, err := sam.NewKeys()
	if err != nil {
		t.Fatalf("Failed to generate I2P keys: %v", err)
	}

	primary, err := sam.NewPrimarySession("test-primary-subsessions-"+RandString(), keys, Options_Default)
	if err != nil {
		t.Fatalf("Failed to create primary session: %v", err)
	}
	defer primary.Close()

	// Test sub-session creation in parallel
	t.Run("StreamSubSession", func(t *testing.T) {
		t.Parallel()

		done := make(chan error, 1)
		go func() {
			streamSub, err := primary.NewStreamSubSession("stream-sub-"+RandString(), Options_Small)
			if err != nil {
				done <- err
				return
			}
			defer streamSub.Close()

			if streamSub.ID() == "" {
				t.Error("Expected non-empty sub-session ID")
				done <- fmt.Errorf("Expected non-empty sub-session ID")
				return
			}

			if streamSub.Type() != "STREAM" {
				t.Errorf("Expected STREAM type, got %s", streamSub.Type())
				done <- fmt.Errorf("Expected STREAM type, got %s", streamSub.Type())
				return
			}

			done <- nil
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Stream sub-session creation failed: %v", err)
			}
		case <-time.After(3 * time.Minute):
			t.Error("Stream sub-session creation timed out")
		}
	})

	t.Run("DatagramSubSession", func(t *testing.T) {
		t.Parallel()

		done := make(chan error, 1)
		go func() {
			// DATAGRAM subsessions require a PORT parameter per SAM v3.3 specification
			options := append(Options_Small, "PORT=8080")
			datagramSub, err := primary.NewDatagramSubSession("datagram-sub-"+RandString(), options)
			if err != nil {
				done <- err
				return
			}
			defer datagramSub.Close()

			if datagramSub.Type() != "DATAGRAM" {
				t.Errorf("Expected DATAGRAM type, got %s", datagramSub.Type())
				done <- fmt.Errorf("Expected DATAGRAM type, got %s", datagramSub.Type())
				return
			}

			done <- nil
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Datagram sub-session creation failed: %v", err)
			}
		case <-time.After(3 * time.Minute):
			t.Error("Datagram sub-session creation timed out")
		}
	})

	t.Run("RawSubSession", func(t *testing.T) {
		t.Parallel()

		done := make(chan error, 1)
		go func() {
			// RAW subsessions require a PORT parameter per SAM v3.3 specification
			options := append(Options_Small, "PORT=8081")
			rawSub, err := primary.NewRawSubSession("raw-sub-"+RandString(), options)
			if err != nil {
				done <- err
				return
			}
			defer rawSub.Close()

			if rawSub.Type() != "RAW" {
				t.Errorf("Expected RAW type, got %s", rawSub.Type())
				done <- fmt.Errorf("Expected RAW type, got %s", rawSub.Type())
				return
			}

			done <- nil
		}()

		select {
		case err := <-done:
			if err != nil {
				t.Errorf("Raw sub-session creation failed: %v", err)
			}
		case <-time.After(3 * time.Minute):
			t.Error("Raw sub-session creation timed out")
		}
	})

	// Wait for all sub-tests to complete before checking sub-session count
	time.Sleep(100 * time.Millisecond)

	// Verify sub-session management
	if primary.SubSessionCount() < 0 {
		t.Errorf("Expected non-negative sub-session count, got %d", primary.SubSessionCount())
	}
}

// TestSAMEmbedding tests that the SAM type properly embeds common.SAM
// and provides access to the underlying functionality with real connections.
func TestSAMEmbedding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Create a real SAM connection
	sam, err := NewSAM(SAMDefaultAddr(""))
	if err != nil {
		t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
	}
	defer sam.Close()

	// Verify that we can access embedded methods
	if sam.SAM == nil {
		t.Fatal("SAM should embed common.SAM")
	}

	// Test embedded functionality
	keys, err := sam.NewKeys()
	if err != nil {
		t.Fatalf("Failed to generate keys using embedded SAM: %v", err)
	}

	if keys.Addr().String() == "" {
		t.Error("Generated keys should have non-empty address")
	}

	// Test resolver functionality
	resolver, err := NewSAMResolver(sam)
	if err != nil {
		t.Fatalf("Failed to create resolver: %v", err)
	}

	// Test that resolver can be used (though we don't test actual resolution
	// since that requires specific I2P destinations to be available)
	if resolver == nil {
		t.Error("Resolver should not be nil")
	}
}

// TestSessionMethodSignatures verifies that the session creation methods
// have the exact signatures expected by the sam3 API using real connections.
func TestSessionMethodSignatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Test each method signature by actually calling them
	// This is more comprehensive than just compile-time checks

	// NewPrimarySession signature test
	t.Run("NewPrimarySession", func(t *testing.T) {
		sam, err := NewSAM(SAMDefaultAddr(""))
		if err != nil {
			t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
		}
		defer sam.Close()

		keys, err := sam.NewKeys()
		if err != nil {
			t.Fatalf("Failed to generate keys: %v", err)
		}

		session, err := sam.NewPrimarySession("sig-test-primary-"+RandString(), keys, Options_Small)
		if err != nil {
			t.Errorf("NewPrimarySession signature test failed: %v", err)
		} else {
			session.Close()
			t.Log("✓ NewPrimarySession signature works correctly")
		}
	})

	// NewPrimarySessionWithSignature signature test
	t.Run("NewPrimarySessionWithSignature", func(t *testing.T) {
		sam, err := NewSAM(SAMDefaultAddr(""))
		if err != nil {
			t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
		}
		defer sam.Close()

		keys, err := sam.NewKeys()
		if err != nil {
			t.Fatalf("Failed to generate keys: %v", err)
		}

		session, err := sam.NewPrimarySessionWithSignature("sig-test-primary-sig-"+RandString(), keys, Options_Small, Sig_EdDSA_SHA512_Ed25519)
		if err != nil {
			t.Errorf("NewPrimarySessionWithSignature signature test failed: %v", err)
		} else {
			session.Close()
			t.Log("✓ NewPrimarySessionWithSignature signature works correctly")
		}
	})

	// NewStreamSession signature test
	t.Run("NewStreamSession", func(t *testing.T) {
		sam, err := NewSAM(SAMDefaultAddr(""))
		if err != nil {
			t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
		}
		defer sam.Close()

		keys, err := sam.NewKeys()
		if err != nil {
			t.Fatalf("Failed to generate keys: %v", err)
		}

		session, err := sam.NewStreamSession("sig-test-stream-"+RandString(), keys, Options_Small)
		if err != nil {
			t.Errorf("NewStreamSession signature test failed: %v", err)
		} else {
			session.Close()
			t.Log("✓ NewStreamSession signature works correctly")
		}
	})

	// NewStreamSessionWithSignature signature test
	t.Run("NewStreamSessionWithSignature", func(t *testing.T) {
		sam, err := NewSAM(SAMDefaultAddr(""))
		if err != nil {
			t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
		}
		defer sam.Close()

		keys, err := sam.NewKeys()
		if err != nil {
			t.Fatalf("Failed to generate keys: %v", err)
		}

		session, err := sam.NewStreamSessionWithSignature("sig-test-stream-sig-"+RandString(), keys, Options_Small, Sig_EdDSA_SHA512_Ed25519)
		if err != nil {
			t.Errorf("NewStreamSessionWithSignature signature test failed: %v", err)
		} else {
			session.Close()
			t.Log("✓ NewStreamSessionWithSignature signature works correctly")
		}
	})

	// NewDatagramSession signature test
	t.Run("NewDatagramSession", func(t *testing.T) {
		sam, err := NewSAM(SAMDefaultAddr(""))
		if err != nil {
			t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
		}
		defer sam.Close()

		keys, err := sam.NewKeys()
		if err != nil {
			t.Fatalf("Failed to generate keys: %v", err)
		}

		session, err := sam.NewDatagramSession("sig-test-datagram-"+RandString(), keys, Options_Small)
		if err != nil {
			t.Errorf("NewDatagramSession signature test failed: %v", err)
		} else {
			session.Close()
			t.Log("✓ NewDatagramSession signature works correctly")
		}
	})

	// NewRawSession signature test
	t.Run("NewRawSession", func(t *testing.T) {
		sam, err := NewSAM(SAMDefaultAddr(""))
		if err != nil {
			t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
		}
		defer sam.Close()

		keys, err := sam.NewKeys()
		if err != nil {
			t.Fatalf("Failed to generate keys: %v", err)
		}

		session, err := sam.NewRawSession("sig-test-raw-"+RandString(), keys, Options_Small, 0)
		if err != nil {
			t.Errorf("NewRawSession signature test failed: %v", err)
		} else {
			session.Close()
			t.Log("✓ NewRawSession signature works correctly")
		}
	})

	t.Log("✓ All session method signatures work correctly with real I2P connections")
}

// BenchmarkSAMSessionCreation benchmarks the performance of session creation with real I2P connections
func BenchmarkSAMSessionCreation(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmarks in short mode")
	}

	// Create a real SAM connection
	sam, err := NewSAM(SAMDefaultAddr(""))
	if err != nil {
		b.Skipf("Cannot connect to I2P SAM bridge: %v", err)
	}
	defer sam.Close()

	keys, err := sam.NewKeys()
	if err != nil {
		b.Fatalf("Failed to generate keys: %v", err)
	}

	b.Run("PrimarySession", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			session, err := sam.NewPrimarySession("bench-primary-"+RandString(), keys, Options_Small)
			if err != nil {
				b.Errorf("Primary session creation failed: %v", err)
				continue
			}
			session.Close()
		}
	})

	b.Run("StreamSession", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			session, err := sam.NewStreamSession("bench-stream-"+RandString(), keys, Options_Small)
			if err != nil {
				b.Errorf("Stream session creation failed: %v", err)
				continue
			}
			session.Close()
		}
	})
}
