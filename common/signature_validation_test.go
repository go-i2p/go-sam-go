package common

import (
	"strings"
	"testing"
)

func TestValidateAndCleanOptions(t *testing.T) {
	tests := []struct {
		name            string
		sigType         string
		options         []string
		expectedOptions []string
		expectWarning   bool
	}{
		{
			name:            "NoSigType_NoConflict",
			sigType:         "",
			options:         []string{"inbound.length=2", "outbound.length=3"},
			expectedOptions: []string{"inbound.length=2", "outbound.length=3"},
			expectWarning:   false,
		},
		{
			name:            "NoSigType_WithSignatureInOptions",
			sigType:         "",
			options:         []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "inbound.length=2"},
			expectedOptions: []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "inbound.length=2"},
			expectWarning:   false,
		},
		{
			name:            "SigType_NoConflict",
			sigType:         "EdDSA_SHA512_Ed25519",
			options:         []string{"inbound.length=2", "outbound.length=3"},
			expectedOptions: []string{"inbound.length=2", "outbound.length=3"},
			expectWarning:   false,
		},
		{
			name:            "SigType_WithConflict_Same",
			sigType:         "EdDSA_SHA512_Ed25519",
			options:         []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "inbound.length=2"},
			expectedOptions: []string{"inbound.length=2"},
			expectWarning:   true,
		},
		{
			name:            "SigType_WithConflict_Different",
			sigType:         "EdDSA_SHA512_Ed25519",
			options:         []string{"SIGNATURE_TYPE=ECDSA_SHA256_P256", "inbound.length=2"},
			expectedOptions: []string{"inbound.length=2"},
			expectWarning:   true,
		},
		{
			name:    "SigType_MultipleConflicts",
			sigType: "EdDSA_SHA512_Ed25519",
			options: []string{
				"SIGNATURE_TYPE=ECDSA_SHA256_P256",
				"inbound.length=2",
				"SIGNATURE_TYPE=DSA_SHA1",
				"outbound.length=3",
			},
			expectedOptions: []string{"inbound.length=2", "outbound.length=3"},
			expectWarning:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateAndCleanOptions(tt.sigType, tt.options)

			// Check that the result matches expected
			if len(result) != len(tt.expectedOptions) {
				t.Errorf("Expected %d options, got %d", len(tt.expectedOptions), len(result))
				t.Errorf("Expected: %v", tt.expectedOptions)
				t.Errorf("Got: %v", result)
				return
			}

			for i, expected := range tt.expectedOptions {
				if result[i] != expected {
					t.Errorf("Option %d: expected %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

func TestValidateSignatureTypeOptions(t *testing.T) {
	tests := []struct {
		name        string
		options     []string
		expectError bool
	}{
		{
			name:        "NoSignatureType",
			options:     []string{"inbound.length=2", "outbound.length=3"},
			expectError: false,
		},
		{
			name:        "SingleSignatureType",
			options:     []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "inbound.length=2"},
			expectError: false,
		},
		{
			name: "MultipleSignatureTypes",
			options: []string{
				"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519",
				"inbound.length=2",
				"SIGNATURE_TYPE=ECDSA_SHA256_P256",
			},
			expectError: true,
		},
		{
			name: "ThreeSignatureTypes",
			options: []string{
				"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519",
				"SIGNATURE_TYPE=ECDSA_SHA256_P256",
				"SIGNATURE_TYPE=DSA_SHA1",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSignatureTypeOptions(tt.options)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestExtractSignatureType(t *testing.T) {
	tests := []struct {
		name              string
		options           []string
		expectedSigType   string
		expectedRemaining []string
	}{
		{
			name:              "NoSignatureType",
			options:           []string{"inbound.length=2", "outbound.length=3"},
			expectedSigType:   "",
			expectedRemaining: []string{"inbound.length=2", "outbound.length=3"},
		},
		{
			name:              "WithSignatureType",
			options:           []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "inbound.length=2"},
			expectedSigType:   "EdDSA_SHA512_Ed25519",
			expectedRemaining: []string{"inbound.length=2"},
		},
		{
			name:              "MultipleSignatureTypes_TakesLast",
			options:           []string{"SIGNATURE_TYPE=ECDSA_SHA256_P256", "inbound.length=2", "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"},
			expectedSigType:   "EdDSA_SHA512_Ed25519",
			expectedRemaining: []string{"inbound.length=2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sigType, remaining := ExtractSignatureType(tt.options)

			if sigType != tt.expectedSigType {
				t.Errorf("Expected signature type %q, got %q", tt.expectedSigType, sigType)
			}

			if len(remaining) != len(tt.expectedRemaining) {
				t.Errorf("Expected %d remaining options, got %d", len(tt.expectedRemaining), len(remaining))
				t.Errorf("Expected: %v", tt.expectedRemaining)
				t.Errorf("Got: %v", remaining)
				return
			}

			for i, expected := range tt.expectedRemaining {
				if remaining[i] != expected {
					t.Errorf("Remaining option %d: expected %q, got %q", i, expected, remaining[i])
				}
			}
		})
	}
}

func TestEnsureSignatureType(t *testing.T) {
	tests := []struct {
		name            string
		sigType         string
		options         []string
		expectedOptions []string
	}{
		{
			name:            "NoSigType_PreserveOptions",
			sigType:         "",
			options:         []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "inbound.length=2"},
			expectedOptions: []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "inbound.length=2"},
		},
		{
			name:            "SigType_RemoveConflicts",
			sigType:         "EdDSA_SHA512_Ed25519",
			options:         []string{"SIGNATURE_TYPE=ECDSA_SHA256_P256", "inbound.length=2"},
			expectedOptions: []string{"inbound.length=2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnsureSignatureType(tt.sigType, tt.options)

			if len(result) != len(tt.expectedOptions) {
				t.Errorf("Expected %d options, got %d", len(tt.expectedOptions), len(result))
				return
			}

			for i, expected := range tt.expectedOptions {
				if result[i] != expected {
					t.Errorf("Option %d: expected %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

// TestSignatureTypeConflictResolution tests the integration of signature type validation
// with the session creation process to ensure conflicts are properly resolved.
func TestSignatureTypeConflictResolution(t *testing.T) {
	t.Run("ConflictResolutionInSessionCreation", func(t *testing.T) {
		// Test signature type configuration
		sigType := "EdDSA_SHA512_Ed25519"

		// Options that would conflict
		conflictingOptions := []string{
			"SIGNATURE_TYPE=ECDSA_SHA256_P256", // This should be removed
			"inbound.length=2",                 // This should be preserved
			"outbound.length=3",                // This should be preserved
		}

		// Test that the validation removes conflicting options
		cleanedOptions := validateAndCleanOptions(sigType, conflictingOptions)

		expectedOptions := []string{"inbound.length=2", "outbound.length=3"}
		if len(cleanedOptions) != len(expectedOptions) {
			t.Errorf("Expected %d cleaned options, got %d", len(expectedOptions), len(cleanedOptions))
			return
		}

		for i, expected := range expectedOptions {
			if cleanedOptions[i] != expected {
				t.Errorf("Cleaned option %d: expected %q, got %q", i, expected, cleanedOptions[i])
			}
		}

		// Ensure no SIGNATURE_TYPE remains in cleaned options
		for _, opt := range cleanedOptions {
			if strings.HasPrefix(opt, "SIGNATURE_TYPE=") {
				t.Errorf("Found unexpected SIGNATURE_TYPE in cleaned options: %q", opt)
			}
		}
	})
}

// TestValidateSubSessionOptions tests that SESSION ADD properly removes
// invalid options according to SAMv3.3 specification.
func TestValidateSubSessionOptions(t *testing.T) {
	tests := []struct {
		name            string
		options         []string
		expectedOptions []string
		expectWarning   bool
	}{
		{
			name:            "ValidOptions",
			options:         []string{"PORT=7000", "FROM_PORT=8080", "TO_PORT=9090"},
			expectedOptions: []string{"PORT=7000", "FROM_PORT=8080", "TO_PORT=9090"},
			expectWarning:   false,
		},
		{
			name:            "WithSignatureType",
			options:         []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "PORT=7000"},
			expectedOptions: []string{"PORT=7000"},
			expectWarning:   true,
		},
		{
			name:            "WithDestination",
			options:         []string{"DESTINATION=TRANSIENT", "PORT=7000"},
			expectedOptions: []string{"PORT=7000"},
			expectWarning:   true,
		},
		{
			name:            "WithLeaseSetEncType",
			options:         []string{"i2cp.leaseSetEncType=4,0", "PORT=7000"},
			expectedOptions: []string{"PORT=7000"},
			expectWarning:   true,
		},
		{
			name:            "WithTunnelOptions",
			options:         []string{"inbound.length=2", "outbound.quantity=3", "PORT=7000"},
			expectedOptions: []string{"PORT=7000"},
			expectWarning:   true,
		},
		{
			name:            "MultipleInvalidTypes",
			options:         []string{"SIGNATURE_TYPE=ECDSA_SHA256_P256", "PORT=7000", "inbound.length=1", "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"},
			expectedOptions: []string{"PORT=7000"},
			expectWarning:   true,
		},
		{
			name:            "AllInvalidOptions",
			options:         []string{"SIGNATURE_TYPE=DSA_SHA1", "DESTINATION=TRANSIENT", "inbound.length=2", "i2cp.leaseSetEncType=4"},
			expectedOptions: []string{},
			expectWarning:   true,
		},
		{
			name:            "EmptyOptions",
			options:         []string{},
			expectedOptions: []string{},
			expectWarning:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateSubSessionOptions(tt.options)

			// Check result length
			if len(result) != len(tt.expectedOptions) {
				t.Errorf("Expected %d options, got %d", len(tt.expectedOptions), len(result))
				t.Errorf("Expected: %v", tt.expectedOptions)
				t.Errorf("Got: %v", result)
				return
			}

			// Check each option
			for i, expected := range tt.expectedOptions {
				if result[i] != expected {
					t.Errorf("Option %d: expected %q, got %q", i, expected, result[i])
				}
			}

			// Ensure no invalid subsession options remain
			invalidPrefixes := []string{"SIGNATURE_TYPE=", "DESTINATION=", "i2cp.leaseSetEncType=", "inbound.", "outbound."}
			for _, opt := range result {
				for _, prefix := range invalidPrefixes {
					if strings.HasPrefix(opt, prefix) {
						t.Errorf("Found unexpected invalid subsession option in result: %q", opt)
					}
				}
			}
		})
	}
}

// TestValidatePrimarySessionOptions tests that SESSION CREATE properly handles
// duplicate options and validates primary session configuration.
func TestValidatePrimarySessionOptions(t *testing.T) {
	tests := []struct {
		name          string
		options       []string
		expectedCount int
		expectWarning bool
	}{
		{
			name:          "NoDuplicates",
			options:       []string{"PORT=7000", "i2cp.leaseSetEncType=4", "inbound.length=2"},
			expectedCount: 3,
			expectWarning: false,
		},
		{
			name:          "DuplicateLeaseSetEncType",
			options:       []string{"i2cp.leaseSetEncType=4", "PORT=7000", "i2cp.leaseSetEncType=4,0"},
			expectedCount: 2, // Should have PORT and one i2cp.leaseSetEncType
			expectWarning: true,
		},
		{
			name:          "DuplicateSignatureType",
			options:       []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "PORT=7000", "SIGNATURE_TYPE=ECDSA_SHA256_P256"},
			expectedCount: 2, // Should have PORT and one SIGNATURE_TYPE
			expectWarning: true,
		},
		{
			name:          "MultipleDuplicates",
			options:       []string{"PORT=7000", "PORT=8000", "i2cp.leaseSetEncType=4", "i2cp.leaseSetEncType=0"},
			expectedCount: 2, // Should have one PORT and one i2cp.leaseSetEncType
			expectWarning: true,
		},
		{
			name:          "OptionsWithoutEquals",
			options:       []string{"SOME_FLAG", "PORT=7000", "ANOTHER_FLAG"},
			expectedCount: 3, // All should be preserved
			expectWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatePrimarySessionOptions(tt.options)

			// Check result count
			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d options, got %d", tt.expectedCount, len(result))
				t.Errorf("Input: %v", tt.options)
				t.Errorf("Result: %v", result)
			}

			// Check for duplicate keys (should not exist after validation)
			seen := make(map[string]bool)
			for _, opt := range result {
				parts := strings.SplitN(opt, "=", 2)
				if len(parts) == 2 {
					key := parts[0]
					if seen[key] {
						t.Errorf("Found duplicate key %q in result: %v", key, result)
					}
					seen[key] = true
				}
			}
		})
	}
}
