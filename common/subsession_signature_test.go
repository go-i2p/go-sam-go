package common

import (
	"strings"
	"testing"
)

// TestAddSubSessionSignatureTypeHandling tests that SESSION ADD correctly handles
// signature type parameters according to SAMv3.3 spec.
// Per spec: "Do not set the DESTINATION option on a SESSION ADD. The subsession
// will use the destination specified in the primary session."
// This also applies to SIGNATURE_TYPE since it's part of the destination.
func TestAddSubSessionSignatureTypeHandling(t *testing.T) {
	tests := []struct {
		name            string
		style           string
		id              string
		options         []string
		expectWarning   bool
		expectedOptions []string
	}{
		{
			name:            "NoSignatureType_Valid",
			style:           "STREAM",
			id:              "test-stream",
			options:         []string{"FROM_PORT=8080", "inbound.length=2"},
			expectWarning:   false,
			expectedOptions: []string{"FROM_PORT=8080", "inbound.length=2"},
		},
		{
			name:            "WithSignatureType_ShouldWarn",
			style:           "DATAGRAM",
			id:              "test-datagram",
			options:         []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "PORT=7000", "inbound.length=2"},
			expectWarning:   true,
			expectedOptions: []string{"PORT=7000", "inbound.length=2"},
		},
		{
			name:            "MultipleSignatureTypes_ShouldWarn",
			style:           "RAW",
			id:              "test-raw",
			options:         []string{"SIGNATURE_TYPE=ECDSA_SHA256_P256", "PORT=7001", "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"},
			expectWarning:   true,
			expectedOptions: []string{"PORT=7001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the buildSessionAddMessage function behavior
			sam := &SAM{}

			// For now, let's just test the current behavior
			message, err := sam.buildSessionAddMessage(tt.style, tt.id, tt.options)
			if err != nil {
				t.Errorf("buildSessionAddMessage() error = %v", err)
				return
			}

			messageStr := string(message)
			t.Logf("Generated message: %s", strings.TrimSpace(messageStr))

			// Check if SIGNATURE_TYPE appears in the message
			hasSignatureType := strings.Contains(messageStr, "SIGNATURE_TYPE=")

			if hasSignatureType {
				t.Logf("⚠️  POTENTIAL ISSUE: SIGNATURE_TYPE found in SESSION ADD message")
				t.Logf("According to SAMv3.3 spec, subsessions should inherit signature type from primary session")

				// Count how many SIGNATURE_TYPE entries
				signatureCount := strings.Count(messageStr, "SIGNATURE_TYPE=")
				if signatureCount > 0 {
					t.Logf("Found %d SIGNATURE_TYPE entries in SESSION ADD message", signatureCount)
					t.Error("SESSION ADD should not contain SIGNATURE_TYPE according to SAMv3.3 spec")
				}
			}

			// Verify expected format
			expectedStart := "SESSION ADD STYLE=" + tt.style + " ID=" + tt.id
			if !strings.HasPrefix(messageStr, expectedStart) {
				t.Errorf("Message should start with %q, got %q", expectedStart, messageStr[:len(expectedStart)])
			}
		})
	}
}

// TestSessionAddVsSessionCreate demonstrates the difference between
// SESSION CREATE (which should allow SIGNATURE_TYPE) and SESSION ADD (which should not).
func TestSessionAddVsSessionCreate(t *testing.T) {
	sam := &SAM{}

	t.Run("SessionCreateWithSignatureType", func(t *testing.T) {
		// SESSION CREATE should allow and handle SIGNATURE_TYPE
		options := []string{"SIGNATURE_TYPE=EdDSA_SHA512_Ed25519", "inbound.length=2"}

		// Mock the SAM config to have a signature type
		sam.SAMEmit.I2PConfig.SigType = "EdDSA_SHA512_Ed25519"

		message, err := sam.buildSessionCreateMessage(options)
		if err != nil {
			t.Errorf("buildSessionCreateMessage() error = %v", err)
			return
		}

		messageStr := string(message)
		t.Logf("SESSION CREATE message: %s", strings.TrimSpace(messageStr))

		// Should handle signature type conflicts properly (this tests our fix)
		if strings.Count(messageStr, "SIGNATURE_TYPE=") > 1 {
			t.Error("SESSION CREATE should not have duplicate SIGNATURE_TYPE entries")
		}
	})

	t.Run("SessionAddWithSignatureType", func(t *testing.T) {
		// SESSION ADD should not include SIGNATURE_TYPE according to spec
		options := []string{"SIGNATURE_TYPE=ECDSA_SHA256_P256", "PORT=7000"}

		message, err := sam.buildSessionAddMessage("DATAGRAM", "test-sub", options)
		if err != nil {
			t.Errorf("buildSessionAddMessage() error = %v", err)
			return
		}

		messageStr := string(message)
		t.Logf("SESSION ADD message: %s", strings.TrimSpace(messageStr))

		// This demonstrates the potential issue
		if strings.Contains(messageStr, "SIGNATURE_TYPE=") {
			t.Error("SESSION ADD should not contain SIGNATURE_TYPE according to SAMv3.3 spec")
			t.Log("ISSUE DETECTED: SESSION ADD allows SIGNATURE_TYPE in options, but spec says subsessions inherit from primary")
		}
	})
}
