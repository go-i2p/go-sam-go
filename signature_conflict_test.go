package sam3

import (
	"strings"
	"testing"

	"github.com/go-i2p/go-sam-go/common"
)

// TestSignatureTypeConflict demonstrates the signature type conflict issue
// and verifies that the solution properly resolves conflicts.
func TestSignatureTypeConflict(t *testing.T) {
	t.Run("ConflictDetection_BEFORE_FIX", func(t *testing.T) {
		// Demonstrate the original issue: signature type specified in both places
		sigTypeParam := "EdDSA_SHA512_Ed25519"                  // Via parameter
		options := []string{"SIGNATURE_TYPE=ECDSA_SHA256_P256"} // Via options

		// This scenario creates ambiguity - which signature type should be used?
		t.Logf("ORIGINAL ISSUE: sigType parameter: %s", sigTypeParam)
		t.Logf("ORIGINAL ISSUE: options with SIGNATURE_TYPE: %v", options)

		// Check if options contains SIGNATURE_TYPE
		hasSignatureInOptions := false
		for _, opt := range options {
			if strings.HasPrefix(opt, "SIGNATURE_TYPE=") {
				hasSignatureInOptions = true
				break
			}
		}

		if hasSignatureInOptions && sigTypeParam != "" {
			t.Log("✓ CONFLICT DETECTED: Signature type specified in both sigType parameter and options")
			t.Log("✓ This demonstrates the original issue that needed fixing")
		}
	})

	t.Run("ConflictResolution_AFTER_FIX", func(t *testing.T) {
		// Demonstrate how the fix resolves conflicts
		sigTypeParam := "EdDSA_SHA512_Ed25519"
		conflictingOptions := []string{
			"SIGNATURE_TYPE=ECDSA_SHA256_P256", // This should be removed
			"inbound.length=2",                 // This should be preserved
			"outbound.length=3",                // This should be preserved
		}

		t.Logf("BEFORE FIX: sigType=%s, options=%v", sigTypeParam, conflictingOptions)

		// Apply the fix using the new validation function
		cleanedOptions := common.EnsureSignatureType(sigTypeParam, conflictingOptions)

		t.Logf("AFTER FIX: sigType=%s, cleanedOptions=%v", sigTypeParam, cleanedOptions)

		// Verify the fix worked correctly
		expectedOptions := []string{"inbound.length=2", "outbound.length=3"}
		if len(cleanedOptions) != len(expectedOptions) {
			t.Errorf("Expected %d options after cleaning, got %d", len(expectedOptions), len(cleanedOptions))
		}

		// Ensure no SIGNATURE_TYPE remains in cleaned options
		for _, opt := range cleanedOptions {
			if strings.HasPrefix(opt, "SIGNATURE_TYPE=") {
				t.Errorf("❌ SIGNATURE_TYPE found in cleaned options: %q", opt)
			}
		}

		// Verify expected options are preserved
		for i, expected := range expectedOptions {
			if i >= len(cleanedOptions) || cleanedOptions[i] != expected {
				t.Errorf("❌ Expected option %q at index %d, got %q", expected, i, cleanedOptions[i])
			}
		}

		t.Log("✅ SUCCESS: Conflicting signature types properly resolved")
		t.Log("✅ sigType parameter took precedence as expected")
		t.Log("✅ Non-signature options were preserved")
	})

	t.Run("MultipleConflictResolution", func(t *testing.T) {
		// Test with multiple SIGNATURE_TYPE entries in options
		sigTypeParam := "EdDSA_SHA512_Ed25519"
		multipleConflictOptions := []string{
			"SIGNATURE_TYPE=ECDSA_SHA256_P256",
			"inbound.length=2",
			"SIGNATURE_TYPE=DSA_SHA1",
			"outbound.length=3",
			"SIGNATURE_TYPE=ECDSA_SHA384_P384",
		}

		t.Logf("BEFORE: Multiple conflicts - sigType=%s", sigTypeParam)
		t.Logf("BEFORE: options=%v", multipleConflictOptions)

		cleanedOptions := common.EnsureSignatureType(sigTypeParam, multipleConflictOptions)

		t.Logf("AFTER: cleanedOptions=%v", cleanedOptions)

		// Should only have the non-signature options
		expectedOptions := []string{"inbound.length=2", "outbound.length=3"}
		if len(cleanedOptions) != len(expectedOptions) {
			t.Errorf("Expected %d options, got %d", len(expectedOptions), len(cleanedOptions))
		}

		// Verify no signature types remain
		for _, opt := range cleanedOptions {
			if strings.HasPrefix(opt, "SIGNATURE_TYPE=") {
				t.Errorf("❌ Unexpected SIGNATURE_TYPE in cleaned options: %q", opt)
			}
		}

		t.Log("✅ SUCCESS: Multiple signature type conflicts resolved")
	})

	t.Run("NoConflict_PreservesOptions", func(t *testing.T) {
		// Test that options without conflicts are preserved
		sigTypeParam := ""
		options := []string{
			"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519",
			"inbound.length=2",
			"outbound.length=3",
		}

		cleanedOptions := common.EnsureSignatureType(sigTypeParam, options)

		// Should preserve all options when no sigType parameter is specified
		if len(cleanedOptions) != len(options) {
			t.Errorf("Expected %d options preserved, got %d", len(options), len(cleanedOptions))
		}

		for i, expected := range options {
			if cleanedOptions[i] != expected {
				t.Errorf("Expected option %q at index %d, got %q", expected, i, cleanedOptions[i])
			}
		}

		t.Log("✅ SUCCESS: Options preserved when no sigType parameter specified")
	})
}
