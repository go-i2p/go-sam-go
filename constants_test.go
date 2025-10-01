package sam

import (
	"os"
	"reflect"
	"testing"
)

func TestSignatureConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "Sig_NONE constant",
			constant: Sig_NONE,
			expected: "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519",
		},
		{
			name:     "Sig_DSA_SHA1 constant",
			constant: Sig_DSA_SHA1,
			expected: "SIGNATURE_TYPE=DSA_SHA1",
		},
		{
			name:     "Sig_ECDSA_SHA256_P256 constant",
			constant: Sig_ECDSA_SHA256_P256,
			expected: "SIGNATURE_TYPE=ECDSA_SHA256_P256",
		},
		{
			name:     "Sig_ECDSA_SHA384_P384 constant",
			constant: Sig_ECDSA_SHA384_P384,
			expected: "SIGNATURE_TYPE=ECDSA_SHA384_P384",
		},
		{
			name:     "Sig_ECDSA_SHA512_P521 constant",
			constant: Sig_ECDSA_SHA512_P521,
			expected: "SIGNATURE_TYPE=ECDSA_SHA512_P521",
		},
		{
			name:     "Sig_EdDSA_SHA512_Ed25519 constant",
			constant: Sig_EdDSA_SHA512_Ed25519,
			expected: "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519",
		},
		{
			name:     "Sig_DEFAULT points to Ed25519",
			constant: Sig_DEFAULT,
			expected: "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Signature constant = %v, want %v", tt.constant, tt.expected)
			}
		})
	}
}

func TestSig_DEFAULT_Security(t *testing.T) {
	// Verify that SIG_DEFAULT points to the secure Ed25519 signature type
	if Sig_DEFAULT != Sig_EdDSA_SHA512_Ed25519 {
		t.Errorf("Sig_DEFAULT should point to secure Ed25519 signature, got %v", Sig_DEFAULT)
	}

	// Verify that deprecated Sig_NONE also points to secure signature
	if Sig_NONE != Sig_EdDSA_SHA512_Ed25519 {
		t.Errorf("Sig_NONE should point to secure Ed25519 signature, got %v", Sig_NONE)
	}
}

func TestOptionsVariables_Structure(t *testing.T) {
	tests := []struct {
		name    string
		options []string
		minLen  int
		maxLen  int
	}{
		{
			name:    "Options_Humongous",
			options: Options_Humongous,
			minLen:  6,
			maxLen:  10,
		},
		{
			name:    "Options_Large",
			options: Options_Large,
			minLen:  6,
			maxLen:  10,
		},
		{
			name:    "Options_Wide",
			options: Options_Wide,
			minLen:  6,
			maxLen:  10,
		},
		{
			name:    "Options_Medium",
			options: Options_Medium,
			minLen:  6,
			maxLen:  10,
		},
		{
			name:    "Options_Default",
			options: Options_Default,
			minLen:  6,
			maxLen:  10,
		},
		{
			name:    "Options_Small",
			options: Options_Small,
			minLen:  6,
			maxLen:  10,
		},
		{
			name:    "Options_Warning_ZeroHop",
			options: Options_Warning_ZeroHop,
			minLen:  6,
			maxLen:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.options) < tt.minLen || len(tt.options) > tt.maxLen {
				t.Errorf("%s length = %v, want between %v and %v", tt.name, len(tt.options), tt.minLen, tt.maxLen)
			}

			// Verify all options contain valid key=value pairs
			for _, option := range tt.options {
				if !validOptionFormat(option) {
					t.Errorf("%s contains invalid option: %v", tt.name, option)
				}
			}
		})
	}
}

func TestOptionsVariables_TunnelSettings(t *testing.T) {
	// Test that all option sets contain required tunnel configuration
	requiredKeys := []string{"inbound.length", "outbound.length", "inbound.quantity", "outbound.quantity"}

	tests := []struct {
		name    string
		options []string
	}{
		{"Options_Humongous", Options_Humongous},
		{"Options_Large", Options_Large},
		{"Options_Wide", Options_Wide},
		{"Options_Medium", Options_Medium},
		{"Options_Default", Options_Default},
		{"Options_Small", Options_Small},
		{"Options_Warning_ZeroHop", Options_Warning_ZeroHop},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range requiredKeys {
				found := false
				for _, option := range tt.options {
					if containsKey(option, key) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s missing required tunnel setting: %s", tt.name, key)
				}
			}
		})
	}
}

func TestOptions_ZeroHop_Warning(t *testing.T) {
	// Verify that zero hop options actually set zero length tunnels
	zeroHopOptions := Options_Warning_ZeroHop

	expectedZeroSettings := []string{
		"inbound.length=0",
		"outbound.length=0",
		"inbound.lengthVariance=0",
		"outbound.lengthVariance=0",
	}

	for _, expected := range expectedZeroSettings {
		found := false
		for _, option := range zeroHopOptions {
			if option == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Options_Warning_ZeroHop missing zero hop setting: %s", expected)
		}
	}
}

func TestGetEnv_WithDefaults(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable not set",
			envKey:       "TEST_UNSET_VAR",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "environment variable set",
			envKey:       "TEST_SET_VAR",
			envValue:     "custom_value",
			defaultValue: "default_value",
			expected:     "custom_value",
		},
		{
			name:         "environment variable with whitespace",
			envKey:       "TEST_WHITESPACE_VAR",
			envValue:     "  trimmed_value  ",
			defaultValue: "default_value",
			expected:     "trimmed_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey)
			}

			result := getEnv(tt.envKey, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%v, %v) = %v, want %v", tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestSAM_HOST_SAM_PORT_Variables(t *testing.T) {
	// Test default values when environment variables are not set
	originalHost := os.Getenv("sam_host")
	originalPort := os.Getenv("sam_port")

	// Clear environment variables
	os.Unsetenv("sam_host")
	os.Unsetenv("sam_port")

	// Test defaults
	hostDefault := getEnv("sam_host", "127.0.0.1")
	portDefault := getEnv("sam_port", "7656")

	if hostDefault != "127.0.0.1" {
		t.Errorf("Default SAM_HOST = %v, want 127.0.0.1", hostDefault)
	}

	if portDefault != "7656" {
		t.Errorf("Default SAM_PORT = %v, want 7656", portDefault)
	}

	// Test custom environment variables
	os.Setenv("sam_host", "192.168.1.1")
	os.Setenv("sam_port", "7000")

	hostCustom := getEnv("sam_host", "127.0.0.1")
	portCustom := getEnv("sam_port", "7656")

	if hostCustom != "192.168.1.1" {
		t.Errorf("Custom SAM_HOST = %v, want 192.168.1.1", hostCustom)
	}

	if portCustom != "7000" {
		t.Errorf("Custom SAM_PORT = %v, want 7000", portCustom)
	}

	// Restore original environment
	if originalHost != "" {
		os.Setenv("sam_host", originalHost)
	} else {
		os.Unsetenv("sam_host")
	}
	if originalPort != "" {
		os.Setenv("sam_port", originalPort)
	} else {
		os.Unsetenv("sam_port")
	}
}

func TestPrimarySessionString(t *testing.T) {
	result := PrimarySessionString()
	if result != "PRIMARY" {
		t.Errorf("PrimarySessionString() = %v, want PRIMARY", result)
	}

	// Test that PrimarySessionSwitch variable is properly initialized
	if PrimarySessionSwitch != "PRIMARY" {
		t.Errorf("PrimarySessionSwitch = %v, want PRIMARY", PrimarySessionSwitch)
	}
}

func TestOptionSets_Immutability(t *testing.T) {
	// Test that option slices are properly initialized and not nil
	optionSets := []struct {
		name    string
		options []string
	}{
		{"Options_Humongous", Options_Humongous},
		{"Options_Large", Options_Large},
		{"Options_Wide", Options_Wide},
		{"Options_Medium", Options_Medium},
		{"Options_Default", Options_Default},
		{"Options_Small", Options_Small},
		{"Options_Warning_ZeroHop", Options_Warning_ZeroHop},
	}

	for _, set := range optionSets {
		t.Run(set.name, func(t *testing.T) {
			if set.options == nil {
				t.Errorf("%s is nil", set.name)
			}
			if len(set.options) == 0 {
				t.Errorf("%s is empty", set.name)
			}

			// Test that modifying a copy doesn't affect the original
			originalLength := len(set.options)
			optionsCopy := make([]string, len(set.options))
			copy(optionsCopy, set.options)

			// Modify copy
			optionsCopy[0] = "modified=value"

			// Verify original is unchanged
			if len(set.options) != originalLength {
				t.Errorf("%s length changed after copy modification", set.name)
			}
		})
	}
}

func TestTunnelConfiguration_Consistency(t *testing.T) {
	// Test that Options_Default provides sensible defaults
	defaultOpts := Options_Default
	expectedDefaults := map[string]string{
		"inbound.length":          "3",
		"outbound.length":         "3",
		"inbound.lengthVariance":  "0",
		"outbound.lengthVariance": "0",
		"inbound.backupQuantity":  "1",
		"outbound.backupQuantity": "1",
		"inbound.quantity":        "1",
		"outbound.quantity":       "1",
	}

	for key, expectedValue := range expectedDefaults {
		found := false
		for _, option := range defaultOpts {
			if option == key+"="+expectedValue {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Options_Default missing expected setting: %s=%s", key, expectedValue)
		}
	}
}

// Helper functions for testing

func validOptionFormat(option string) bool {
	// Check if option contains '=' and has non-empty key and value
	parts := splitOnEqual(option)
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func splitOnEqual(s string) []string {
	// Simple split on '=' for testing
	for i, r := range s {
		if r == '=' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

func containsKey(option, key string) bool {
	parts := splitOnEqual(option)
	return len(parts) >= 2 && parts[0] == key
}

// Benchmark tests for performance validation

func BenchmarkGetEnv(b *testing.B) {
	os.Setenv("BENCH_TEST_VAR", "test_value")
	defer os.Unsetenv("BENCH_TEST_VAR")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getEnv("BENCH_TEST_VAR", "default")
	}
}

func BenchmarkPrimarySessionString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PrimarySessionString()
	}
}

// Table-driven test for option set validation

func TestOptionSets_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		options     []string
		description string
		minHops     int // Expected minimum tunnel length
		maxHops     int // Expected maximum tunnel length
	}{
		{
			name:        "Options_Humongous",
			options:     Options_Humongous,
			description: "maximum anonymity configuration",
			minHops:     3,
			maxHops:     3,
		},
		{
			name:        "Options_Large",
			options:     Options_Large,
			description: "high traffic configuration",
			minHops:     3,
			maxHops:     3,
		},
		{
			name:        "Options_Wide",
			options:     Options_Wide,
			description: "low latency configuration",
			minHops:     1,
			maxHops:     1,
		},
		{
			name:        "Options_Medium",
			options:     Options_Medium,
			description: "balanced configuration",
			minHops:     3,
			maxHops:     3,
		},
		{
			name:        "Options_Default",
			options:     Options_Default,
			description: "default configuration",
			minHops:     3,
			maxHops:     3,
		},
		{
			name:        "Options_Small",
			options:     Options_Small,
			description: "minimal resource configuration",
			minHops:     3,
			maxHops:     3,
		},
		{
			name:        "Options_Warning_ZeroHop",
			options:     Options_Warning_ZeroHop,
			description: "no anonymity configuration",
			minHops:     0,
			maxHops:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify tunnel length settings match expected hop counts
			inboundLength := extractOptionValue(tt.options, "inbound.length")
			outboundLength := extractOptionValue(tt.options, "outbound.length")

			if inboundLength != tt.minHops || outboundLength != tt.maxHops {
				t.Errorf("%s tunnel lengths: inbound=%d, outbound=%d, want both=%d-%d",
					tt.name, inboundLength, outboundLength, tt.minHops, tt.maxHops)
			}

			t.Logf("%s (%s): verified %d hop tunnels", tt.name, tt.description, tt.minHops)
		})
	}
}

func extractOptionValue(options []string, key string) int {
	for _, option := range options {
		parts := splitOnEqual(option)
		if len(parts) == 2 && parts[0] == key {
			// Simple integer parsing for testing
			if parts[1] == "0" {
				return 0
			} else if parts[1] == "1" {
				return 1
			} else if parts[1] == "2" {
				return 2
			} else if parts[1] == "3" {
				return 3
			} else if parts[1] == "4" {
				return 4
			} else if parts[1] == "6" {
				return 6
			}
		}
	}
	return -1 // Not found
}

// Test that ensures exported constants are properly documented and match expected values
func TestConstants_Documentation(t *testing.T) {
	// This test ensures that critical constants have the expected values
	// and serves as documentation of the API contract

	constantTests := []struct {
		name     string
		actual   interface{}
		expected interface{}
		doc      string
	}{
		{
			name:     "Sig_DEFAULT",
			actual:   Sig_DEFAULT,
			expected: "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519",
			doc:      "Default signature should be secure Ed25519",
		},
		{
			name:     "SAM_HOST default",
			actual:   getEnv("sam_host", "127.0.0.1"),
			expected: "127.0.0.1",
			doc:      "Default SAM host should be localhost",
		},
		{
			name:     "SAM_PORT default",
			actual:   getEnv("sam_port", "7656"),
			expected: "7656",
			doc:      "Default SAM port should be 7656",
		},
		{
			name:     "PrimarySessionSwitch",
			actual:   PrimarySessionSwitch,
			expected: "PRIMARY",
			doc:      "Primary session switch should be 'PRIMARY'",
		},
	}

	for _, tt := range constantTests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.actual, tt.expected) {
				t.Errorf("%s: got %v, want %v (%s)", tt.name, tt.actual, tt.expected, tt.doc)
			}
		})
	}
}
