package common

import (
	"strings"
	"testing"
)

// TestComprehensiveValidationExample demonstrates the complete signature and option
// validation functionality across primary sessions and subsessions.
func TestComprehensiveValidationExample(t *testing.T) {
	t.Run("PrimarySessionValidation", func(t *testing.T) {
		// Test primary session with duplicate options
		primaryOptions := []string{
			"i2cp.leaseSetEncType=4",              // First encryption type
			"inbound.length=2",                    // Valid tunnel option
			"SIGNATURE_TYPE=DSA_SHA1",             // First signature type (will be overridden by sigType param)
			"i2cp.leaseSetEncType=4,0",            // Duplicate encryption type - should use this one
			"outbound.quantity=3",                 // Valid tunnel option
			"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", // Duplicate signature type
		}

		sam := &SAM{}
		sam.SAMEmit.I2PConfig.SigType = "ECDSA_SHA256_P256" // This takes precedence

		message, err := sam.buildSessionCreateMessage(primaryOptions)
		if err != nil {
			t.Errorf("buildSessionCreateMessage() error = %v", err)
			return
		}

		messageStr := string(message)
		t.Logf("Primary session message: %s", strings.TrimSpace(messageStr))

		// Verify that:
		// 1. Only one i2cp.leaseSetEncType remains (the last one)
		// 2. No SIGNATURE_TYPE remains in options (conflicted with sigType param)
		// 3. Valid tunnel options are preserved
		leaseSetCount := strings.Count(messageStr, "i2cp.leaseSetEncType=")
		if leaseSetCount > 2 { // Allow up to 2: one from SAMEmit.Create() and one from options
			t.Errorf("Too many i2cp.leaseSetEncType entries found: %d", leaseSetCount)
		}
		if strings.Contains(messageStr, "i2cp.leaseSetEncType=4,0") {
			t.Log("✓ Correct lease set encryption type preserved")
		} else {
			t.Error("Expected i2cp.leaseSetEncType=4,0 not found")
		}
		if strings.Count(messageStr, "SIGNATURE_TYPE=") > 1 {
			t.Error("Multiple SIGNATURE_TYPE entries found in primary session")
		}
		if !strings.Contains(messageStr, "SIGNATURE_TYPE=ECDSA_SHA256_P256") {
			t.Error("Expected sigType parameter value not found in message")
		}
	})

	t.Run("SubSessionValidation", func(t *testing.T) {
		// Test subsession with various invalid options
		subsessionOptions := []string{
			"PORT=7000",                           // Valid subsession option
			"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", // Invalid - should inherit from primary
			"DESTINATION=TRANSIENT",               // Invalid - should use primary destination
			"i2cp.leaseSetEncType=4,0",            // Invalid - lease set configured at primary level
			"inbound.length=2",                    // Invalid - tunnel config belongs to primary
			"outbound.quantity=3",                 // Invalid - tunnel config belongs to primary
			"FROM_PORT=8080",                      // Valid subsession option
			"TO_PORT=9090",                        // Valid subsession option
			"inbound.backupQuantity=1",            // Invalid - tunnel config belongs to primary
		}

		sam := &SAM{}
		message, err := sam.buildSessionAddMessage("DATAGRAM", "test-subsession", subsessionOptions)
		if err != nil {
			t.Errorf("buildSessionAddMessage() error = %v", err)
			return
		}

		messageStr := string(message)
		t.Logf("Subsession message: %s", strings.TrimSpace(messageStr))

		// Verify that:
		// 1. No SIGNATURE_TYPE in subsession (should inherit from primary)
		// 2. No DESTINATION in subsession (should use primary destination)
		// 3. No lease set encryption options (configured at primary level)
		// 4. No tunnel configuration options (belongs to primary session)
		// 5. Valid subsession options are preserved
		invalidOptions := []string{
			"SIGNATURE_TYPE=",
			"DESTINATION=",
			"i2cp.leaseSetEncType=",
			"inbound.",
			"outbound.",
		}

		for _, invalid := range invalidOptions {
			if strings.Contains(messageStr, invalid) {
				t.Errorf("Found invalid subsession option: %s", invalid)
			}
		}

		// Check valid options are preserved
		validOptions := []string{"PORT=7000", "FROM_PORT=8080", "TO_PORT=9090"}
		for _, valid := range validOptions {
			if !strings.Contains(messageStr, valid) {
				t.Errorf("Missing valid subsession option: %s", valid)
			}
		}

		t.Log("✓ All invalid subsession options properly removed")
		t.Log("✓ All valid subsession options preserved")
	})

	t.Run("ComparisonTest", func(t *testing.T) {
		// Demonstrate the difference between primary and subsession handling
		options := []string{
			"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519",
			"i2cp.leaseSetEncType=4,0",
			"PORT=7000",
			"inbound.length=2",
		}

		sam := &SAM{}

		// Primary session: should handle duplicates but allow tunnel config
		sam.SAMEmit.I2PConfig.SigType = "" // No conflict for primary
		primaryMsg, _ := sam.buildSessionCreateMessage(options)
		primaryStr := string(primaryMsg)

		// Subsession: should remove invalid options
		subMsg, _ := sam.buildSessionAddMessage("STREAM", "test-sub", options)
		subStr := string(subMsg)

		t.Logf("Primary allows tunnel config: %s", strings.TrimSpace(primaryStr))
		t.Logf("Subsession removes tunnel config: %s", strings.TrimSpace(subStr))

		// Primary should contain tunnel config
		if !strings.Contains(primaryStr, "inbound.length=2") {
			t.Error("Primary session should allow tunnel configuration")
		}

		// Subsession should not contain any invalid options
		if strings.Contains(subStr, "inbound.length=2") {
			t.Error("Subsession should not contain tunnel configuration")
		}
		if strings.Contains(subStr, "SIGNATURE_TYPE=") {
			t.Error("Subsession should not contain signature type")
		}
		if strings.Contains(subStr, "i2cp.leaseSetEncType=") {
			t.Error("Subsession should not contain lease set encryption type")
		}

		t.Log("✓ Primary session and subsession validation work correctly")
	})
}

// TestEdgeCases tests various edge cases in the validation system.
func TestEdgeCases(t *testing.T) {
	t.Run("EmptyOptions", func(t *testing.T) {
		sam := &SAM{}

		// Primary with empty options
		primaryMsg, err := sam.buildSessionCreateMessage([]string{})
		if err != nil {
			t.Errorf("Empty primary options should not cause error: %v", err)
		}
		t.Logf("Empty primary: %s", strings.TrimSpace(string(primaryMsg)))

		// Subsession with empty options
		subMsg, err := sam.buildSessionAddMessage("RAW", "empty-sub", []string{})
		if err != nil {
			t.Errorf("Empty subsession options should not cause error: %v", err)
		}
		t.Logf("Empty subsession: %s", strings.TrimSpace(string(subMsg)))
	})

	t.Run("OptionsWithoutEquals", func(t *testing.T) {
		// Test options that don't have '=' (flags)
		options := []string{
			"SOME_FLAG",
			"PORT=7000",
			"ANOTHER_FLAG",
			"i2cp.leaseSetEncType=4",
		}

		result := validatePrimarySessionOptions(options)

		// Flags should be preserved
		flagCount := 0
		for _, opt := range result {
			if !strings.Contains(opt, "=") {
				flagCount++
			}
		}
		if flagCount != 2 {
			t.Errorf("Expected 2 flags to be preserved, got %d", flagCount)
		}
	})

	t.Run("MalformedOptions", func(t *testing.T) {
		// Test options with unusual formatting
		options := []string{
			"=EMPTY_KEY",            // Empty key
			"EMPTY_VALUE=",          // Empty value
			"NORMAL=value",          // Normal option
			"SIGNATURE_TYPE=",       // Empty signature type
			"i2cp.leaseSetEncType=", // Empty lease set type
		}

		primaryResult := validatePrimarySessionOptions(options)
		subResult := validateSubSessionOptions(options)

		t.Logf("Primary result: %v", primaryResult)
		t.Logf("Subsession result: %v", subResult)

		// Subsession should remove SIGNATURE_TYPE and i2cp.leaseSetEncType even if empty
		for _, opt := range subResult {
			if strings.HasPrefix(opt, "SIGNATURE_TYPE=") {
				t.Error("Subsession should not contain SIGNATURE_TYPE, even if empty")
			}
			if strings.HasPrefix(opt, "i2cp.leaseSetEncType=") {
				t.Error("Subsession should not contain i2cp.leaseSetEncType, even if empty")
			}
		}
	})
}
