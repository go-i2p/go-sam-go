package sam

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/samber/oops"
	"github.com/sirupsen/logrus"
)

// Test data constants for consistent testing
const (
	testSAMAddress       = "127.0.0.1:7656"
	testDestination      = "ABCD1234567890abcdef0123456789ABCDEF012345678901234567890abcdef0123456789ABCDEF01234567890123456789"
	testResponseWithDest = testDestination + " RESULT=OK MESSAGE=Session created"
	testKeyValueString   = "HOST=example.org PORT=1234 TYPE=stream STATUS=active"
	testOptions          = "inbound.length=3 outbound.length=3 inbound.quantity=2 outbound.quantity=2"
)

// TestNewSAM tests the main SAM constructor function
func TestNewSAM(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		wantError bool
		errorType string
	}{
		{
			name:      "empty address",
			address:   "",
			wantError: true,
			errorType: "connection",
		},
		{
			name:      "invalid address format",
			address:   "invalid:address:format:extra",
			wantError: true,
			errorType: "connection",
		},
		{
			name:      "non-existent host",
			address:   "nonexistent.invalid:7656",
			wantError: true,
			errorType: "connection",
		},
		{
			name:      "invalid port number",
			address:   "127.0.0.1:99999",
			wantError: true,
			errorType: "connection",
		},
		{
			name:      "valid address format but unreachable",
			address:   "127.0.0.1:7656",
			wantError: false, // Function will succeed in creating SAM instance, connection may fail later
			errorType: "connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sam, err := NewSAM(tt.address)
			if tt.wantError {
				if err == nil {
					t.Errorf("NewSAM() expected error but got none")
				}
				if sam != nil {
					t.Errorf("NewSAM() expected nil SAM instance on error")
				}
			} else {
				if err != nil {
					t.Errorf("NewSAM() unexpected error: %v", err)
				}
				if sam == nil {
					t.Errorf("NewSAM() expected SAM instance but got nil")
				} else {
					sam.Close()
				}
			}
		})
	}
}

// TestExtractDest tests destination extraction from SAM responses
func TestExtractDest(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "valid response with destination",
			input:  testResponseWithDest,
			output: testDestination,
		},
		{
			name:   "single word input",
			input:  "DESTINATION_ONLY",
			output: "DESTINATION_ONLY",
		},
		{
			name:   "empty input",
			input:  "",
			output: "",
		},
		{
			name:   "whitespace only",
			input:  "   ",
			output: "",
		},
		{
			name:   "multiple spaces",
			input:  "DEST   RESULT=OK   STATUS=active",
			output: "DEST",
		},
		{
			name:   "newline in input",
			input:  "DEST\nRESTOFLINE",
			output: "DEST\nRESTOFLINE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractDest(tt.input)
			if result != tt.output {
				t.Errorf("ExtractDest(%q) = %q, want %q", tt.input, result, tt.output)
			}
		})
	}
}

// TestExtractPairInt tests integer value extraction from key-value pairs
func TestExtractPairInt(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		key    string
		output int
	}{
		{
			name:   "valid port extraction",
			input:  testKeyValueString,
			key:    "PORT",
			output: 1234,
		},
		{
			name:   "non-existent key",
			input:  testKeyValueString,
			key:    "NONEXISTENT",
			output: 0,
		},
		{
			name:   "empty input",
			input:  "",
			key:    "PORT",
			output: 0,
		},
		{
			name:   "invalid integer value",
			input:  "PORT=invalid TYPE=stream",
			key:    "PORT",
			output: 0,
		},
		{
			name:   "negative integer",
			input:  "PORT=-1234 TYPE=stream",
			key:    "PORT",
			output: -1234,
		},
		{
			name:   "zero value",
			input:  "PORT=0 TYPE=stream",
			key:    "PORT",
			output: 0,
		},
		{
			name:   "large integer",
			input:  "PORT=65535 TYPE=stream",
			key:    "PORT",
			output: 65535,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPairInt(tt.input, tt.key)
			if result != tt.output {
				t.Errorf("ExtractPairInt(%q, %q) = %d, want %d", tt.input, tt.key, result, tt.output)
			}
		})
	}
}

// TestExtractPairString tests string value extraction from key-value pairs
func TestExtractPairString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		key    string
		output string
	}{
		{
			name:   "valid host extraction",
			input:  testKeyValueString,
			key:    "HOST",
			output: "example.org",
		},
		{
			name:   "valid type extraction",
			input:  testKeyValueString,
			key:    "TYPE",
			output: "stream",
		},
		{
			name:   "non-existent key",
			input:  testKeyValueString,
			key:    "NONEXISTENT",
			output: "",
		},
		{
			name:   "empty input",
			input:  "",
			key:    "HOST",
			output: "",
		},
		{
			name:   "key without value",
			input:  "HOST= PORT=1234",
			key:    "HOST",
			output: "",
		},
		{
			name:   "key with spaces in value",
			input:  "MESSAGE=hello_world STATUS=ok",
			key:    "MESSAGE",
			output: "hello_world",
		},
		{
			name:   "duplicate keys",
			input:  "HOST=first.org HOST=second.org",
			key:    "HOST",
			output: "first.org", // Should return first occurrence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPairString(tt.input, tt.key)
			if result != tt.output {
				t.Errorf("ExtractPairString(%q, %q) = %q, want %q", tt.input, tt.key, result, tt.output)
			}
		})
	}
}

// TestGenerateOptionString tests option string generation from slices
func TestGenerateOptionString(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output string
	}{
		{
			name:   "multiple options",
			input:  []string{"inbound.length=3", "outbound.length=3", "inbound.quantity=2"},
			output: "inbound.length=3 outbound.length=3 inbound.quantity=2",
		},
		{
			name:   "single option",
			input:  []string{"inbound.length=3"},
			output: "inbound.length=3",
		},
		{
			name:   "empty slice",
			input:  []string{},
			output: "",
		},
		{
			name:   "nil slice",
			input:  nil,
			output: "",
		},
		{
			name:   "options with spaces",
			input:  []string{"option with spaces", "another=value"},
			output: "option with spaces another=value",
		},
		{
			name:   "empty option in slice",
			input:  []string{"valid=option", "", "another=value"},
			output: "valid=option  another=value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateOptionString(tt.input)
			if result != tt.output {
				t.Errorf("GenerateOptionString(%v) = %q, want %q", tt.input, result, tt.output)
			}
		})
	}
}

// TestGetSAM3Logger tests logger initialization and configuration
func TestGetSAM3Logger(t *testing.T) {
	logger := GetSAM3Logger()

	// Verify logger is not nil
	if logger == nil {
		t.Fatal("GetSAM3Logger() returned nil")
	}

	// Verify logger is a logrus instance (it should be since that's what we return)
	if logger.Level > logrus.InfoLevel {
		t.Errorf("GetSAM3Logger() logger level is %v, expected at least Info level", logger.Level)
	}

	// Test logger functionality
	logger.Info("Test log message")
}

// TestIgnorePortError tests port error filtering functionality
func TestIgnorePortError(t *testing.T) {
	tests := []struct {
		name      string
		input     error
		shouldNil bool
	}{
		{
			name:      "nil error",
			input:     nil,
			shouldNil: true,
		},
		{
			name:      "missing port error",
			input:     errors.New("missing port in address"),
			shouldNil: true,
		},
		{
			name:      "other network error",
			input:     errors.New("connection refused"),
			shouldNil: false,
		},
		{
			name:      "wrapped missing port error",
			input:     oops.Errorf("parsing failed: missing port in address"),
			shouldNil: true,
		},
		{
			name:      "partial match should not be ignored",
			input:     errors.New("missing something else"),
			shouldNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IgnorePortError(tt.input)
			if tt.shouldNil && result != nil {
				t.Errorf("IgnorePortError(%v) should return nil but got %v", tt.input, result)
			}
			if !tt.shouldNil && result == nil {
				t.Errorf("IgnorePortError(%v) should preserve error but got nil", tt.input)
			}
			if !tt.shouldNil && result != tt.input {
				t.Errorf("IgnorePortError(%v) should return original error unchanged", tt.input)
			}
		})
	}
}

// TestInitializeSAM3Logger tests logger initialization function
func TestInitializeSAM3Logger(t *testing.T) {
	// This is a simple test since the function primarily performs setup
	// In a real implementation, we might check log configuration

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("InitializeSAM3Logger() panicked: %v", r)
		}
	}()

	InitializeSAM3Logger()

	// Verify we can still get a logger after initialization
	logger := GetSAM3Logger()
	if logger == nil {
		t.Error("GetSAM3Logger() returned nil after InitializeSAM3Logger()")
	}
}

// TestRandString tests random string generation
func TestRandString(t *testing.T) {
	// Test basic functionality
	result1 := RandString()
	result2 := RandString()

	// Verify non-empty results
	if result1 == "" {
		t.Error("RandString() returned empty string")
	}
	if result2 == "" {
		t.Error("RandString() returned empty string on second call")
	}

	// Verify results are different (extremely high probability)
	if result1 == result2 {
		t.Error("RandString() returned identical strings (very unlikely)")
	}

	// Verify expected length
	expectedLength := 12
	if len(result1) != expectedLength {
		t.Errorf("RandString() returned string of length %d, expected %d", len(result1), expectedLength)
	}

	// Verify character set (alphanumeric lowercase)
	validChars := "abcdefghijklmnopqrstuvwxyz0123456789"
	for _, char := range result1 {
		if !strings.ContainsRune(validChars, char) {
			t.Errorf("RandString() returned string with invalid character: %c", char)
		}
	}

	// Test multiple calls for consistency
	for i := 0; i < 10; i++ {
		result := RandString()
		if len(result) != expectedLength {
			t.Errorf("RandString() call %d returned string of length %d, expected %d", i, len(result), expectedLength)
		}
	}
}

// TestSAMDefaultAddr tests SAM address construction with fallback
func TestSAMDefaultAddr(t *testing.T) {
	tests := []struct {
		name       string
		fallback   string
		expectAddr string
	}{
		{
			name:       "with fallback",
			fallback:   "192.168.1.100:7656",
			expectAddr: "127.0.0.1:7656", // Should use SAM_HOST:SAM_PORT from constants
		},
		{
			name:       "empty fallback",
			fallback:   "",
			expectAddr: "127.0.0.1:7656", // Should still use SAM_HOST:SAM_PORT
		},
		{
			name:       "different fallback port",
			fallback:   "localhost:9999",
			expectAddr: "127.0.0.1:7656", // Should use SAM_HOST:SAM_PORT, not fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SAMDefaultAddr(tt.fallback)
			if result != tt.expectAddr {
				t.Errorf("SAMDefaultAddr(%q) = %q, want %q", tt.fallback, result, tt.expectAddr)
			}
		})
	}
}

// TestSplitHostPort tests I2P-aware host/port splitting
func TestSplitHostPort(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectHost  string
		expectPort  string
		expectError bool
	}{
		{
			name:        "standard host:port",
			input:       "example.com:8080",
			expectHost:  "example.com",
			expectPort:  "8080",
			expectError: false,
		},
		{
			name:        "I2P address without port",
			input:       "example.i2p",
			expectHost:  "example.i2p",
			expectPort:  "0",
			expectError: false,
		},
		{
			name:        "localhost with port",
			input:       "localhost:7656",
			expectHost:  "localhost",
			expectPort:  "7656",
			expectError: false,
		},
		{
			name:        "IPv6 address with port",
			input:       "[::1]:7656",
			expectHost:  "::1",
			expectPort:  "7656",
			expectError: false,
		},
		{
			name:        "empty input",
			input:       "",
			expectHost:  "",
			expectPort:  "0",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, err := SplitHostPort(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("SplitHostPort(%q) expected error but got none", tt.input)
			}
			if !tt.expectError && err != nil {
				t.Errorf("SplitHostPort(%q) unexpected error: %v", tt.input, err)
			}
			if host != tt.expectHost {
				t.Errorf("SplitHostPort(%q) host = %q, want %q", tt.input, host, tt.expectHost)
			}
			if port != tt.expectPort {
				t.Errorf("SplitHostPort(%q) port = %q, want %q", tt.input, port, tt.expectPort)
			}
		})
	}
}

// TestNewSAMResolver tests SAM resolver creation
func TestNewSAMResolver(t *testing.T) {
	// Test with nil SAM instance
	t.Run("nil SAM instance", func(t *testing.T) {
		resolver, err := NewSAMResolver(nil)
		// The common package might handle nil gracefully or panic
		// We test actual behavior rather than expectations
		if err != nil && resolver != nil {
			t.Error("NewSAMResolver(nil) returned both error and non-nil resolver")
		}
		// Note: The actual behavior depends on common.NewSAMResolver implementation
	})
}

// TestNewFullSAMResolver tests complete SAM resolver creation
func TestNewFullSAMResolver(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		wantError bool
	}{
		{
			name:      "empty address",
			address:   "",
			wantError: true,
		},
		{
			name:      "invalid address",
			address:   "invalid:address:format",
			wantError: true,
		},
		{
			name:      "unreachable address",
			address:   "127.0.0.1:7656",
			wantError: false, // Should succeed in creating resolver even if no SAM bridge
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver, err := NewFullSAMResolver(tt.address)
			if tt.wantError {
				if err == nil {
					t.Errorf("NewFullSAMResolver(%q) expected error but got none", tt.address)
				}
				if resolver != nil {
					t.Errorf("NewFullSAMResolver(%q) expected nil resolver on error", tt.address)
				}
			} else {
				if err != nil {
					t.Errorf("NewFullSAMResolver(%q) unexpected error: %v", tt.address, err)
				}
				if resolver == nil {
					t.Errorf("NewFullSAMResolver(%q) expected resolver but got nil", tt.address)
				} else {
					// In a real implementation, we'd call resolver.Close()
				}
			}
		})
	}
}

// TestDelegationFunctions tests that wrapper functions properly delegate to common package
func TestDelegationFunctions(t *testing.T) {
	// Test that utility functions match their common package counterparts
	t.Run("ExtractDest delegation", func(t *testing.T) {
		input := "DEST RESULT=OK"
		samResult := ExtractDest(input)
		commonResult := common.ExtractDest(input)
		if samResult != commonResult {
			t.Errorf("ExtractDest delegation failed: sam=%q, common=%q", samResult, commonResult)
		}
	})

	t.Run("ExtractPairString delegation", func(t *testing.T) {
		input := "HOST=example.org PORT=1234"
		key := "HOST"
		samResult := ExtractPairString(input, key)
		commonResult := common.ExtractPairString(input, key)
		if samResult != commonResult {
			t.Errorf("ExtractPairString delegation failed: sam=%q, common=%q", samResult, commonResult)
		}
	})

	t.Run("ExtractPairInt delegation", func(t *testing.T) {
		input := "HOST=example.org PORT=1234"
		key := "PORT"
		samResult := ExtractPairInt(input, key)
		commonResult := common.ExtractPairInt(input, key)
		if samResult != commonResult {
			t.Errorf("ExtractPairInt delegation failed: sam=%d, common=%d", samResult, commonResult)
		}
	})

	t.Run("IgnorePortError delegation", func(t *testing.T) {
		testErr := errors.New("missing port in address")
		samResult := IgnorePortError(testErr)
		commonResult := common.IgnorePortError(testErr)
		if (samResult == nil) != (commonResult == nil) {
			t.Errorf("IgnorePortError delegation failed: sam=%v, common=%v", samResult, commonResult)
		}
	})

	t.Run("SplitHostPort delegation", func(t *testing.T) {
		input := "example.com:8080"
		samHost, samPort, samErr := SplitHostPort(input)
		commonHost, commonPort, commonErr := common.SplitHostPort(input)

		if samHost != commonHost || samPort != commonPort {
			t.Errorf("SplitHostPort delegation failed: sam=(%q,%q), common=(%q,%q)",
				samHost, samPort, commonHost, commonPort)
		}
		if (samErr == nil) != (commonErr == nil) {
			t.Errorf("SplitHostPort error delegation failed: sam=%v, common=%v", samErr, commonErr)
		}
	})
}

// BenchmarkUtilityFunctions provides performance benchmarks for utility functions
func BenchmarkUtilityFunctions(b *testing.B) {
	testInput := "HOST=example.org PORT=1234 TYPE=stream STATUS=active DEST=" + testDestination

	b.Run("ExtractDest", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ExtractDest(testInput)
		}
	})

	b.Run("ExtractPairString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ExtractPairString(testInput, "HOST")
		}
	})

	b.Run("ExtractPairInt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ExtractPairInt(testInput, "PORT")
		}
	})

	b.Run("GenerateOptionString", func(b *testing.B) {
		options := []string{"inbound.length=3", "outbound.length=3", "inbound.quantity=2", "outbound.quantity=2"}
		for i := 0; i < b.N; i++ {
			GenerateOptionString(options)
		}
	})

	b.Run("RandString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RandString()
		}
	})
}
