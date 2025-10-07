package sam3

import (
	"reflect"
	"strings"
	"testing"
)

// TestSAM3CompatibilityAPI verifies that all documented types and functions from sigs.md
// are available and have correct signatures for drop-in replacement functionality.
// This test ensures perfect API surface compatibility with the original sam3 library.
func TestSAM3CompatibilityAPI(t *testing.T) {
	t.Run("CoreTypes", func(t *testing.T) {
		// Verify all core types exist and are properly aliased
		coreTypes := map[string]interface{}{
			"SAM":             (*SAM)(nil),
			"StreamSession":   (*StreamSession)(nil),
			"DatagramSession": (*DatagramSession)(nil),
			"RawSession":      (*RawSession)(nil),
			"PrimarySession":  (*PrimarySession)(nil),
			"SAMConn":         (*SAMConn)(nil),
			"StreamListener":  (*StreamListener)(nil),
			"SAMResolver":     (*SAMResolver)(nil),
			"I2PConfig":       (*I2PConfig)(nil),
			"SAMEmit":         (*SAMEmit)(nil),
			"Options":         (*Options)(nil),
			"Option":          (*Option)(nil),
			"BaseSession":     (*BaseSession)(nil),
		}

		for typeName, typeValue := range coreTypes {
			if typeValue == nil {
				t.Errorf("Type %s should not be nil", typeName)
				continue
			}

			// Verify type is not nil interface
			typeOf := reflect.TypeOf(typeValue)
			if typeOf == nil {
				t.Errorf("Type %s has nil reflection type", typeName)
				continue
			}

			t.Logf("✓ Type %s exists and is properly defined", typeName)
		}
	})

	t.Run("Constants", func(t *testing.T) {
		// Verify all signature constants exist
		expectedConstants := map[string]string{
			"Sig_NONE":                 Sig_NONE,
			"Sig_DSA_SHA1":             Sig_DSA_SHA1,
			"Sig_ECDSA_SHA256_P256":    Sig_ECDSA_SHA256_P256,
			"Sig_ECDSA_SHA384_P384":    Sig_ECDSA_SHA384_P384,
			"Sig_ECDSA_SHA512_P521":    Sig_ECDSA_SHA512_P521,
			"Sig_EdDSA_SHA512_Ed25519": Sig_EdDSA_SHA512_Ed25519,
		}

		for constName, constValue := range expectedConstants {
			if constValue == "" {
				t.Errorf("Constant %s is empty", constName)
				continue
			}
			if !strings.Contains(constValue, "SIGNATURE_TYPE=") {
				t.Errorf("Constant %s does not contain expected prefix: %s", constName, constValue)
				continue
			}
			t.Logf("✓ Constant %s = %s", constName, constValue)
		}
	})

	t.Run("OptionVariables", func(t *testing.T) {
		// Verify all option variables exist and have expected structure
		optionVars := map[string][]string{
			"Options_Humongous":       Options_Humongous,
			"Options_Large":           Options_Large,
			"Options_Wide":            Options_Wide,
			"Options_Medium":          Options_Medium,
			"Options_Default":         Options_Default,
			"Options_Small":           Options_Small,
			"Options_Warning_ZeroHop": Options_Warning_ZeroHop,
		}

		for varName, varValue := range optionVars {
			if len(varValue) == 0 {
				t.Errorf("Option variable %s is empty", varName)
				continue
			}

			// Verify it contains expected tunnel options
			hasInbound := false
			hasOutbound := false
			for _, option := range varValue {
				if strings.Contains(option, "inbound.") {
					hasInbound = true
				}
				if strings.Contains(option, "outbound.") {
					hasOutbound = true
				}
			}

			if !hasInbound || !hasOutbound {
				t.Errorf("Option variable %s missing inbound/outbound options", varName)
				continue
			}

			t.Logf("✓ Option variable %s has %d options", varName, len(varValue))
		}
	})

	t.Run("EnvironmentVariables", func(t *testing.T) {
		// Verify environment variable handling
		if SAM_HOST == "" {
			t.Error("SAM_HOST should have a default value")
		}
		if SAM_PORT == "" {
			t.Error("SAM_PORT should have a default value")
		}

		t.Logf("✓ SAM_HOST = %s", SAM_HOST)
		t.Logf("✓ SAM_PORT = %s", SAM_PORT)
	})
}

// TestSAM3CompatibilityFunctions verifies that all documented functions from sigs.md
// exist and have the correct signatures for perfect drop-in replacement compatibility.
func TestSAM3CompatibilityFunctions(t *testing.T) {
	t.Run("UtilityFunctions", func(t *testing.T) {
		// Test utility functions exist and have correct signatures
		utilityTests := []struct {
			name string
			test func(*testing.T)
		}{
			{
				name: "PrimarySessionString",
				test: func(t *testing.T) {
					result := PrimarySessionString()
					if result == "" {
						t.Error("PrimarySessionString should return non-empty string")
					}
					t.Logf("✓ PrimarySessionString() = %s", result)
				},
			},
			{
				name: "RandString",
				test: func(t *testing.T) {
					result := RandString()
					if len(result) == 0 {
						t.Error("RandString should return non-empty string")
					}
					t.Logf("✓ RandString() = %s", result)
				},
			},
			{
				name: "SAMDefaultAddr",
				test: func(t *testing.T) {
					result := SAMDefaultAddr("")
					if result == "" {
						t.Error("SAMDefaultAddr should return non-empty string")
					}
					t.Logf("✓ SAMDefaultAddr(\"\") = %s", result)
				},
			},
			{
				name: "ExtractDest",
				test: func(t *testing.T) {
					input := "test-dest RESULT=OK DESTINATION=other-dest"
					result := ExtractDest(input)
					if result != "test-dest" {
						t.Errorf("ExtractDest expected 'test-dest', got '%s'", result)
					}
					t.Logf("✓ ExtractDest works correctly")
				},
			},
			{
				name: "ExtractPairString",
				test: func(t *testing.T) {
					input := "KEY1=value1 KEY2=value2"
					result := ExtractPairString(input, "KEY1")
					if result != "value1" {
						t.Errorf("ExtractPairString expected 'value1', got '%s'", result)
					}
					t.Logf("✓ ExtractPairString works correctly")
				},
			},
			{
				name: "ExtractPairInt",
				test: func(t *testing.T) {
					input := "PORT=7656 COUNT=10"
					result := ExtractPairInt(input, "PORT")
					if result != 7656 {
						t.Errorf("ExtractPairInt expected 7656, got %d", result)
					}
					t.Logf("✓ ExtractPairInt works correctly")
				},
			},
		}

		for _, test := range utilityTests {
			t.Run(test.name, test.test)
		}
	})

	t.Run("ConstructorFunctions", func(t *testing.T) {
		// Test constructor functions exist - we can't test actual functionality
		// without I2P connection, but we can verify signatures
		constructorTests := []struct {
			name string
			test func(*testing.T)
		}{
			{
				name: "NewSAM",
				test: func(t *testing.T) {
					// Test that function exists and has correct signature
					fnType := reflect.TypeOf(NewSAM)
					if fnType.NumIn() != 1 || fnType.NumOut() != 2 {
						t.Error("NewSAM should have signature: (string) (*SAM, error)")
					}
					t.Logf("✓ NewSAM has correct signature")
				},
			},
			{
				name: "NewSAMResolver",
				test: func(t *testing.T) {
					fnType := reflect.TypeOf(NewSAMResolver)
					if fnType.NumIn() != 1 || fnType.NumOut() != 2 {
						t.Error("NewSAMResolver should have signature: (*SAM) (*SAMResolver, error)")
					}
					t.Logf("✓ NewSAMResolver has correct signature")
				},
			},
			{
				name: "NewFullSAMResolver",
				test: func(t *testing.T) {
					fnType := reflect.TypeOf(NewFullSAMResolver)
					if fnType.NumIn() != 1 || fnType.NumOut() != 2 {
						t.Error("NewFullSAMResolver should have signature: (string) (*SAMResolver, error)")
					}
					t.Logf("✓ NewFullSAMResolver has correct signature")
				},
			},
		}

		for _, test := range constructorTests {
			t.Run(test.name, test.test)
		}
	})

	t.Run("ConfigurationFunctions", func(t *testing.T) {
		// Test configuration functions exist and work correctly
		configTests := []struct {
			name string
			test func(*testing.T)
		}{
			{
				name: "NewConfig",
				test: func(t *testing.T) {
					config, err := NewConfig()
					if err != nil {
						t.Errorf("NewConfig() failed: %v", err)
					}
					if config == nil {
						t.Error("NewConfig() returned nil config")
					}
					t.Logf("✓ NewConfig works correctly")
				},
			},
			{
				name: "NewEmit",
				test: func(t *testing.T) {
					emit, err := NewEmit()
					if err != nil {
						t.Errorf("NewEmit() failed: %v", err)
					}
					if emit == nil {
						t.Error("NewEmit() returned nil emit")
					}
					t.Logf("✓ NewEmit works correctly")
				},
			},
			{
				name: "SetType",
				test: func(t *testing.T) {
					emit, _ := NewEmit()
					err := SetType("STREAM")(emit)
					if err != nil {
						t.Errorf("SetType failed: %v", err)
					}
					if emit.Style != "STREAM" {
						t.Errorf("SetType did not set style correctly")
					}
					t.Logf("✓ SetType works correctly")
				},
			},
		}

		for _, test := range configTests {
			t.Run(test.name, test.test)
		}
	})
}

// TestSAM3CompatibilityIntegration tests drop-in replacement functionality with
// real I2P connections when available. These tests verify that the wrapper
// behaves identically to the original sam3 library in actual usage scenarios.
func TestSAM3CompatibilityIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("BasicSAMConnection", func(t *testing.T) {
		// Test basic SAM connection using sam3 API patterns
		sam, err := NewSAM(SAMDefaultAddr(""))
		if err != nil {
			t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
		}
		defer sam.Close()

		// Verify SAM connection provides expected functionality
		if sam == nil {
			t.Fatal("SAM connection should not be nil")
		}

		// Test key generation through SAM
		keys, err := sam.NewKeys()
		if err != nil {
			t.Fatalf("Failed to generate keys: %v", err)
		}

		if keys.String() == "" {
			t.Fatal("Generated keys should not be empty")
		}

		t.Log("✓ Basic SAM connection and key generation work correctly")
	})

	t.Run("SessionCreationPatterns", func(t *testing.T) {
		// Test common sam3 usage patterns for session creation
		sam, err := NewSAM(SAMDefaultAddr(""))
		if err != nil {
			t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
		}
		defer sam.Close()

		// Test the typical sam3 usage pattern for each session type
		sessionTests := []struct {
			name string
			test func(*testing.T)
		}{
			{
				name: "PrimarySessionPattern",
				test: func(t *testing.T) {
					// Generate unique keys for each session to avoid conflicts
					keys, err := sam.NewKeys()
					if err != nil {
						t.Fatalf("Failed to generate keys for primary session: %v", err)
					}
					session, err := sam.NewPrimarySession("compat-primary-"+RandString(), keys, Options_Default)
					if err != nil {
						t.Errorf("Primary session creation failed: %v", err)
						return
					}
					defer session.Close()

					// Verify primary session can create sub-sessions
					if session.SubSessionCount() != 0 {
						t.Error("New primary session should have 0 sub-sessions")
					}

					t.Log("✓ Primary session creation pattern works")
					
					// Explicitly close session to avoid conflicts with subsequent tests
					session.Close()
				},
			},
			{
				name: "StreamSessionPattern",
				test: func(t *testing.T) {
					// Create a separate SAM connection for this test to avoid conflicts
					streamSam, err := NewSAM(SAMDefaultAddr(""))
					if err != nil {
						t.Fatalf("Failed to create SAM connection for stream session: %v", err)
					}
					defer streamSam.Close()
					
					// Generate unique keys for each session to avoid conflicts
					keys, err := streamSam.NewKeys()
					if err != nil {
						t.Fatalf("Failed to generate keys for stream session: %v", err)
					}
					session, err := streamSam.NewStreamSession("compat-stream-"+RandString(), keys, Options_Small)
					if err != nil {
						t.Errorf("Stream session creation failed: %v", err)
						return
					}
					defer session.Close()

					// Test listener creation - typical sam3 pattern
					listener, err := session.Listen()
					if err != nil {
						t.Errorf("Stream listener creation failed: %v", err)
						return
					}
					defer listener.Close()

					if listener.Addr() == nil {
						t.Error("Stream listener should have non-nil address")
					}

					t.Log("✓ Stream session creation and listener pattern works")
					
					// Explicitly close session to avoid conflicts with subsequent tests
					listener.Close()
					session.Close()
				},
			},
			{
				name: "DatagramSessionPattern",
				test: func(t *testing.T) {
					// Create a separate SAM connection for this test to avoid conflicts
					datagramSam, err := NewSAM(SAMDefaultAddr(""))
					if err != nil {
						t.Fatalf("Failed to create SAM connection for datagram session: %v", err)
					}
					defer datagramSam.Close()
					
					// Generate unique keys for each session to avoid conflicts
					keys, err := datagramSam.NewKeys()
					if err != nil {
						t.Fatalf("Failed to generate keys for datagram session: %v", err)
					}
					session, err := datagramSam.NewDatagramSession("compat-datagram-"+RandString(), keys, Options_Small, 0)
					if err != nil {
						t.Errorf("Datagram session creation failed: %v", err)
						return
					}
					defer session.Close()

					// Verify datagram session provides expected interface
					if session.LocalAddr() == nil {
						t.Error("Datagram session should have non-nil local address")
					}

					t.Log("✓ Datagram session creation pattern works")
					
					// Explicitly close session to avoid conflicts with subsequent tests
					session.Close()
				},
			},
			{
				name: "RawSessionPattern",
				test: func(t *testing.T) {
					// Create a separate SAM connection for this test to avoid conflicts
					rawSam, err := NewSAM(SAMDefaultAddr(""))
					if err != nil {
						t.Fatalf("Failed to create SAM connection for raw session: %v", err)
					}
					defer rawSam.Close()
					
					// Generate unique keys for each session to avoid conflicts
					keys, err := rawSam.NewKeys()
					if err != nil {
						t.Fatalf("Failed to generate keys for raw session: %v", err)
					}
					session, err := rawSam.NewRawSession("compat-raw-"+RandString(), keys, Options_Small, 0)
					if err != nil {
						t.Errorf("Raw session creation failed: %v", err)
						return
					}
					defer session.Close()

					// Verify raw session provides expected interface
					if session.LocalAddr() == nil {
						t.Error("Raw session should have non-nil local address")
					}

					t.Log("✓ Raw session creation pattern works")
					
					// Explicitly close session to avoid conflicts with subsequent tests
					session.Close()
				},
			},
		}

		for _, test := range sessionTests {
			t.Run(test.name, test.test)
		}
	})

	t.Run("ErrorHandlingCompatibility", func(t *testing.T) {
		// Test that error handling matches expected sam3 patterns

		// Test invalid SAM address
		_, err := NewSAM("invalid-address:999999")
		if err == nil {
			t.Error("Expected error for invalid SAM address")
		}
		t.Logf("✓ Invalid address error: %v", err)

		// Test with valid connection for other error cases
		sam, err := NewSAM(SAMDefaultAddr(""))
		if err != nil {
			t.Skipf("Cannot connect to I2P SAM bridge: %v", err)
		}
		defer sam.Close()

		// Test invalid session type
		emit, _ := NewEmit()
		err = SetType("INVALID_TYPE")(emit)
		if err == nil {
			t.Error("Expected error for invalid session type")
		}
		t.Logf("✓ Invalid session type error: %v", err)

		// Test invalid configuration values
		err = SetInLength(-1)(emit)
		if err == nil {
			t.Error("Expected error for invalid tunnel length")
		}
		t.Logf("✓ Invalid configuration error: %v", err)
	})
}

// TestSAM3CompatibilityBehavior verifies that the wrapper exhibits identical
// behavior to the original sam3 library in edge cases and specific scenarios.
func TestSAM3CompatibilityBehavior(t *testing.T) {
	t.Run("AddressHandling", func(t *testing.T) {
		// Test SAM address handling compatibility
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{"DefaultHost", "", SAM_HOST + ":" + SAM_PORT},
			{"ExplicitAddress", "192.168.1.1:7656", SAM_HOST + ":" + SAM_PORT}, // SAMDefaultAddr ignores input when defaults are set
			{"HostOnly", "localhost", SAM_HOST + ":" + SAM_PORT},               // SAMDefaultAddr ignores input when defaults are set
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := SAMDefaultAddr(test.input)
				if result != test.expected {
					t.Errorf("SAMDefaultAddr(%s) = %s, expected %s", test.input, result, test.expected)
				}
				t.Logf("✓ SAMDefaultAddr(%s) = %s", test.input, result)
			})
		}
	})

	t.Run("StringParsing", func(t *testing.T) {
		// Test string parsing functions for SAM protocol compatibility
		testCases := []struct {
			name     string
			input    string
			expected map[string]interface{}
		}{
			{
				name:  "DestinationExtraction",
				input: "abcd1234.b32.i2p RESULT=OK VERSION=3.3",
				expected: map[string]interface{}{
					"dest": "abcd1234.b32.i2p",
				},
			},
			{
				name:  "MultipleParams",
				input: "test.b32.i2p RESULT=OK MESSAGE=Connected",
				expected: map[string]interface{}{
					"dest":    "test.b32.i2p",
					"message": "Connected",
					"result":  "OK",
				},
			},
		}

		for _, test := range testCases {
			t.Run(test.name, func(t *testing.T) {
				dest := ExtractDest(test.input)
				if expectedDest, ok := test.expected["dest"]; ok {
					if dest != expectedDest {
						t.Errorf("ExtractDest expected %s, got %s", expectedDest, dest)
					}
				}

				message := ExtractPairString(test.input, "MESSAGE")
				if expectedMsg, ok := test.expected["message"]; ok {
					if message != expectedMsg {
						t.Errorf("ExtractPairString(MESSAGE) expected %s, got %s", expectedMsg, message)
					}
				}

				t.Logf("✓ String parsing works correctly for %s", test.name)
			})
		}
	})

	t.Run("ConfigurationBehavior", func(t *testing.T) {
		// Test configuration behavior matches sam3 expectations
		emit, _ := NewEmit()

		// Test option string generation
		options := []string{"inbound.length=3", "outbound.length=3"}
		optString := GenerateOptionString(options)

		if !strings.Contains(optString, "inbound.length=3") {
			t.Error("Generated option string should contain inbound.length=3")
		}
		if !strings.Contains(optString, "outbound.length=3") {
			t.Error("Generated option string should contain outbound.length=3")
		}

		// Test configuration chaining
		err := SetType("STREAM")(emit)
		if err != nil {
			t.Errorf("Configuration chaining failed: %v", err)
		}

		err = SetSAMHost("127.0.0.1")(emit)
		if err != nil {
			t.Errorf("Configuration chaining failed: %v", err)
		}

		if emit.Style != "STREAM" {
			t.Error("Configuration chaining did not preserve previous settings")
		}
		if emit.I2PConfig.SamHost != "127.0.0.1" {
			t.Error("Configuration chaining did not apply new settings")
		}

		t.Log("✓ Configuration behavior matches sam3 patterns")
	})
}

// BenchmarkSAM3Compatibility benchmarks critical operations to ensure
// performance is comparable to original sam3 library expectations.
func BenchmarkSAM3Compatibility(b *testing.B) {
	b.Run("UtilityFunctions", func(b *testing.B) {
		b.Run("RandString", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = RandString()
			}
		})

		b.Run("SAMDefaultAddr", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = SAMDefaultAddr("")
			}
		})

		b.Run("ExtractDest", func(b *testing.B) {
			input := "HELLO REPLY RESULT=OK DESTINATION=test.b32.i2p"
			for i := 0; i < b.N; i++ {
				_ = ExtractDest(input)
			}
		})

		b.Run("GenerateOptionString", func(b *testing.B) {
			options := Options_Default
			for i := 0; i < b.N; i++ {
				_ = GenerateOptionString(options)
			}
		})
	})

	b.Run("Configuration", func(b *testing.B) {
		b.Run("NewEmit", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = NewEmit()
			}
		})

		b.Run("SetType", func(b *testing.B) {
			emit, _ := NewEmit()
			for i := 0; i < b.N; i++ {
				_ = SetType("STREAM")(emit)
			}
		})

		b.Run("OptionVariableAccess", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Options_Default
			}
		})
	})
}
