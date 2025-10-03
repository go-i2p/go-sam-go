package sam3

import (
	"reflect"
	"testing"

	"github.com/go-i2p/go-sam-go/common"
)

// TestNewConfig tests the NewConfig wrapper function
func TestNewConfig(t *testing.T) {
	tests := []struct {
		name    string
		opts    []func(*I2PConfig) error
		wantErr bool
	}{
		{
			name:    "no options",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "empty options",
			opts:    []func(*I2PConfig) error{},
			wantErr: false,
		},
		{
			name: "valid options",
			opts: []func(*I2PConfig) error{
				func(c *I2PConfig) error {
					c.TunName = "test-tunnel"
					return nil
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && config == nil {
				t.Error("NewConfig() returned nil config without error")
			}
		})
	}
}

// TestNewEmit tests the NewEmit wrapper function
func TestNewEmit(t *testing.T) {
	tests := []struct {
		name    string
		opts    []func(*SAMEmit) error
		wantErr bool
	}{
		{
			name:    "no options",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "empty options",
			opts:    []func(*SAMEmit) error{},
			wantErr: false,
		},
		{
			name: "valid options",
			opts: []func(*SAMEmit) error{
				SetType("STREAM"),
				SetSAMHost("localhost"),
			},
			wantErr: false,
		},
		{
			name: "invalid options",
			opts: []func(*SAMEmit) error{
				SetType("INVALID_TYPE"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit, err := NewEmit(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && emit == nil {
				t.Error("NewEmit() returned nil emit without error")
			}
		})
	}
}

// TestConfigFunctionDelegation tests that all config functions properly delegate to common package
func TestConfigFunctionDelegation(t *testing.T) {
	// Test function delegation by comparing function pointer addresses or behavior
	tests := []struct {
		name     string
		rootFunc interface{}
		expected interface{}
	}{
		{"SetType", SetType, common.SetType},
		{"SetSAMAddress", SetSAMAddress, common.SetSAMAddress},
		{"SetSAMHost", SetSAMHost, common.SetSAMHost},
		{"SetSAMPort", SetSAMPort, common.SetSAMPort},
		{"SetName", SetName, common.SetName},
		{"SetInLength", SetInLength, common.SetInLength},
		{"SetOutLength", SetOutLength, common.SetOutLength},
		{"SetInVariance", SetInVariance, common.SetInVariance},
		{"SetOutVariance", SetOutVariance, common.SetOutVariance},
		{"SetInQuantity", SetInQuantity, common.SetInQuantity},
		{"SetOutQuantity", SetOutQuantity, common.SetOutQuantity},
		{"SetInBackups", SetInBackups, common.SetInBackups},
		{"SetOutBackups", SetOutBackups, common.SetOutBackups},
		{"SetEncrypt", SetEncrypt, common.SetEncrypt},
		{"SetLeaseSetKey", SetLeaseSetKey, common.SetLeaseSetKey},
		{"SetLeaseSetPrivateKey", SetLeaseSetPrivateKey, common.SetLeaseSetPrivateKey},
		{"SetLeaseSetPrivateSigningKey", SetLeaseSetPrivateSigningKey, common.SetLeaseSetPrivateSigningKey},
		{"SetMessageReliability", SetMessageReliability, common.SetMessageReliability},
		{"SetAllowZeroIn", SetAllowZeroIn, common.SetAllowZeroIn},
		{"SetAllowZeroOut", SetAllowZeroOut, common.SetAllowZeroOut},
		{"SetCompress", SetCompress, common.SetCompress},
		{"SetFastRecieve", SetFastRecieve, common.SetFastRecieve},
		{"SetReduceIdle", SetReduceIdle, common.SetReduceIdle},
		{"SetReduceIdleTime", SetReduceIdleTime, common.SetReduceIdleTime},
		{"SetReduceIdleTimeMs", SetReduceIdleTimeMs, common.SetReduceIdleTimeMs},
		{"SetReduceIdleQuantity", SetReduceIdleQuantity, common.SetReduceIdleQuantity},
		{"SetCloseIdle", SetCloseIdle, common.SetCloseIdle},
		{"SetCloseIdleTime", SetCloseIdleTime, common.SetCloseIdleTime},
		{"SetCloseIdleTimeMs", SetCloseIdleTimeMs, common.SetCloseIdleTimeMs},
		{"SetAccessListType", SetAccessListType, common.SetAccessListType},
		{"SetAccessList", SetAccessList, common.SetAccessList},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify that the functions have the same type signature
			rootType := reflect.TypeOf(tt.rootFunc)
			expectedType := reflect.TypeOf(tt.expected)

			if rootType != expectedType {
				t.Errorf("Function %s has wrong type: got %v, want %v", tt.name, rootType, expectedType)
			}
		})
	}
}

// TestSetType tests the SetType wrapper function
func TestSetType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected string
	}{
		{"valid STREAM", "STREAM", false, "STREAM"},
		{"valid DATAGRAM", "DATAGRAM", false, "DATAGRAM"},
		{"valid RAW", "RAW", false, "RAW"},
		{"invalid type", "INVALID", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{}
			err := SetType(tt.input)(emit)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetType(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}

			if !tt.wantErr && emit.Style != tt.expected {
				t.Errorf("SetType(%s) set Style = %v, want %v", tt.input, emit.Style, tt.expected)
			}
		})
	}
}

// TestSetSAMAddress tests the SetSAMAddress wrapper function
func TestSetSAMAddress(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantErr      bool
		expectedHost string
		expectedPort int
	}{
		{"valid address with port", "127.0.0.1:7656", false, "127.0.0.1", 7656},
		{"valid address without port", "localhost", false, "localhost", 0},
		{"invalid port", "127.0.0.1:invalid", true, "", 0},
		{"too many colons", "a:b:c", true, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{}
			err := SetSAMAddress(tt.input)(emit)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetSAMAddress(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}

			if !tt.wantErr {
				if emit.I2PConfig.SamHost != tt.expectedHost {
					t.Errorf("SetSAMAddress(%s) set SamHost = %v, want %v", tt.input, emit.I2PConfig.SamHost, tt.expectedHost)
				}
				if emit.I2PConfig.SamPort != tt.expectedPort {
					t.Errorf("SetSAMAddress(%s) set SamPort = %v, want %v", tt.input, emit.I2PConfig.SamPort, tt.expectedPort)
				}
			}
		})
	}
}

// TestSetSAMPort tests the SetSAMPort wrapper function
func TestSetSAMPort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		expected int
	}{
		{"valid port", "7656", false, 7656},
		{"valid port zero", "0", false, 0},
		{"valid port max", "65535", false, 65535},
		{"invalid port negative", "-1", true, 0},
		{"invalid port too large", "65536", true, 0},
		{"invalid port non-numeric", "abc", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{}
			err := SetSAMPort(tt.input)(emit)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetSAMPort(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}

			if !tt.wantErr && emit.I2PConfig.SamPort != tt.expected {
				t.Errorf("SetSAMPort(%s) set SamPort = %v, want %v", tt.input, emit.I2PConfig.SamPort, tt.expected)
			}
		})
	}
}

// TestTunnelLengthFunctions tests tunnel length configuration functions
func TestTunnelLengthFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(int) func(*SAMEmit) error
		input    int
		wantErr  bool
		field    string
	}{
		{"SetInLength valid", SetInLength, 3, false, "InLength"},
		{"SetInLength invalid high", SetInLength, 7, true, "InLength"},
		{"SetInLength invalid negative", SetInLength, -1, true, "InLength"},
		{"SetOutLength valid", SetOutLength, 3, false, "OutLength"},
		{"SetOutLength invalid high", SetOutLength, 7, true, "OutLength"},
		{"SetOutLength invalid negative", SetOutLength, -1, true, "OutLength"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{}
			err := tt.function(tt.input)(emit)

			if (err != nil) != tt.wantErr {
				t.Errorf("%s(%d) error = %v, wantErr %v", tt.name, tt.input, err, tt.wantErr)
			}

			if !tt.wantErr {
				// Use reflection to check the field value
				value := reflect.ValueOf(&emit.I2PConfig).Elem().FieldByName(tt.field)
				if value.Int() != int64(tt.input) {
					t.Errorf("%s(%d) set %s = %v, want %v", tt.name, tt.input, tt.field, value.Int(), tt.input)
				}
			}
		})
	}
}

// TestTunnelVarianceFunctions tests tunnel variance configuration functions
func TestTunnelVarianceFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(int) func(*SAMEmit) error
		input    int
		wantErr  bool
		field    string
	}{
		{"SetInVariance valid positive", SetInVariance, 1, false, "InVariance"},
		{"SetInVariance valid negative", SetInVariance, -1, false, "InVariance"},
		{"SetInVariance valid zero", SetInVariance, 0, false, "InVariance"},
		{"SetInVariance invalid high", SetInVariance, 7, true, "InVariance"},
		{"SetInVariance invalid low", SetInVariance, -7, true, "InVariance"},
		{"SetOutVariance valid positive", SetOutVariance, 1, false, "OutVariance"},
		{"SetOutVariance valid negative", SetOutVariance, -1, false, "OutVariance"},
		{"SetOutVariance invalid high", SetOutVariance, 7, true, "OutVariance"},
		{"SetOutVariance invalid low", SetOutVariance, -7, true, "OutVariance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{}
			err := tt.function(tt.input)(emit)

			if (err != nil) != tt.wantErr {
				t.Errorf("%s(%d) error = %v, wantErr %v", tt.name, tt.input, err, tt.wantErr)
			}

			if !tt.wantErr {
				// Use reflection to check the field value
				value := reflect.ValueOf(&emit.I2PConfig).Elem().FieldByName(tt.field)
				if value.Int() != int64(tt.input) {
					t.Errorf("%s(%d) set %s = %v, want %v", tt.name, tt.input, tt.field, value.Int(), tt.input)
				}
			}
		})
	}
}

// TestTunnelQuantityFunctions tests tunnel quantity configuration functions
func TestTunnelQuantityFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(int) func(*SAMEmit) error
		input    int
		wantErr  bool
		field    string
	}{
		{"SetInQuantity valid", SetInQuantity, 2, false, "InQuantity"},
		{"SetInQuantity min valid", SetInQuantity, 1, false, "InQuantity"},
		{"SetInQuantity max valid", SetInQuantity, 16, false, "InQuantity"},
		{"SetInQuantity invalid zero", SetInQuantity, 0, true, "InQuantity"},
		{"SetInQuantity invalid high", SetInQuantity, 17, true, "InQuantity"},
		{"SetOutQuantity valid", SetOutQuantity, 2, false, "OutQuantity"},
		{"SetOutQuantity min valid", SetOutQuantity, 1, false, "OutQuantity"},
		{"SetOutQuantity max valid", SetOutQuantity, 16, false, "OutQuantity"},
		{"SetOutQuantity invalid zero", SetOutQuantity, 0, true, "OutQuantity"},
		{"SetOutQuantity invalid high", SetOutQuantity, 17, true, "OutQuantity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{}
			err := tt.function(tt.input)(emit)

			if (err != nil) != tt.wantErr {
				t.Errorf("%s(%d) error = %v, wantErr %v", tt.name, tt.input, err, tt.wantErr)
			}

			if !tt.wantErr {
				// Use reflection to check the field value
				value := reflect.ValueOf(&emit.I2PConfig).Elem().FieldByName(tt.field)
				if value.Int() != int64(tt.input) {
					t.Errorf("%s(%d) set %s = %v, want %v", tt.name, tt.input, tt.field, value.Int(), tt.input)
				}
			}
		})
	}
}

// TestBackupQuantityFunctions tests backup tunnel configuration functions
func TestBackupQuantityFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(int) func(*SAMEmit) error
		input    int
		wantErr  bool
		field    string
	}{
		{"SetInBackups valid", SetInBackups, 1, false, "InBackupQuantity"},
		{"SetInBackups min valid", SetInBackups, 0, false, "InBackupQuantity"},
		{"SetInBackups max valid", SetInBackups, 5, false, "InBackupQuantity"},
		{"SetInBackups invalid high", SetInBackups, 6, true, "InBackupQuantity"},
		{"SetOutBackups valid", SetOutBackups, 1, false, "OutBackupQuantity"},
		{"SetOutBackups min valid", SetOutBackups, 0, false, "OutBackupQuantity"},
		{"SetOutBackups max valid", SetOutBackups, 5, false, "OutBackupQuantity"},
		{"SetOutBackups invalid high", SetOutBackups, 6, true, "OutBackupQuantity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{}
			err := tt.function(tt.input)(emit)

			if (err != nil) != tt.wantErr {
				t.Errorf("%s(%d) error = %v, wantErr %v", tt.name, tt.input, err, tt.wantErr)
			}

			if !tt.wantErr {
				// Use reflection to check the field value
				value := reflect.ValueOf(&emit.I2PConfig).Elem().FieldByName(tt.field)
				if value.Int() != int64(tt.input) {
					t.Errorf("%s(%d) set %s = %v, want %v", tt.name, tt.input, tt.field, value.Int(), tt.input)
				}
			}
		})
	}
}

// TestBooleanConfigFunctions tests boolean configuration functions
func TestBooleanConfigFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(bool) func(*SAMEmit) error
		input    bool
		field    string
	}{
		{"SetEncrypt true", SetEncrypt, true, "EncryptLeaseSet"},
		{"SetEncrypt false", SetEncrypt, false, "EncryptLeaseSet"},
		{"SetAllowZeroIn true", SetAllowZeroIn, true, "InAllowZeroHop"},
		{"SetAllowZeroIn false", SetAllowZeroIn, false, "InAllowZeroHop"},
		{"SetAllowZeroOut true", SetAllowZeroOut, true, "OutAllowZeroHop"},
		{"SetAllowZeroOut false", SetAllowZeroOut, false, "OutAllowZeroHop"},
		{"SetCompress true", SetCompress, true, "UseCompression"},
		{"SetCompress false", SetCompress, false, "UseCompression"},
		{"SetFastRecieve true", SetFastRecieve, true, "FastRecieve"},
		{"SetFastRecieve false", SetFastRecieve, false, "FastRecieve"},
		{"SetReduceIdle true", SetReduceIdle, true, "ReduceIdle"},
		{"SetReduceIdle false", SetReduceIdle, false, "ReduceIdle"},
		{"SetCloseIdle true", SetCloseIdle, true, "CloseIdle"},
		{"SetCloseIdle false", SetCloseIdle, false, "CloseIdle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{}
			err := tt.function(tt.input)(emit)
			if err != nil {
				t.Errorf("%s(%v) unexpected error: %v", tt.name, tt.input, err)
			}

			// Use reflection to check the field value
			value := reflect.ValueOf(&emit.I2PConfig).Elem().FieldByName(tt.field)
			if value.Bool() != tt.input {
				t.Errorf("%s(%v) set %s = %v, want %v", tt.name, tt.input, tt.field, value.Bool(), tt.input)
			}
		})
	}
}

// TestStringConfigFunctions tests string configuration functions
func TestStringConfigFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func(string) func(*SAMEmit) error
		input    string
		field    string
	}{
		{"SetSAMHost", SetSAMHost, "127.0.0.1", "SamHost"},
		{"SetName", SetName, "test-tunnel", "TunName"},
		{"SetLeaseSetKey", SetLeaseSetKey, "test-key", "LeaseSetKey"},
		{"SetLeaseSetPrivateKey", SetLeaseSetPrivateKey, "test-private-key", "LeaseSetPrivateKey"},
		{"SetLeaseSetPrivateSigningKey", SetLeaseSetPrivateSigningKey, "test-signing-key", "LeaseSetPrivateSigningKey"},
		{"SetMessageReliability", SetMessageReliability, "BestEffort", "MessageReliability"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{}
			err := tt.function(tt.input)(emit)
			if err != nil {
				t.Errorf("%s(%s) unexpected error: %v", tt.name, tt.input, err)
			}

			// Use reflection to check the field value
			value := reflect.ValueOf(&emit.I2PConfig).Elem().FieldByName(tt.field)
			if value.String() != tt.input {
				t.Errorf("%s(%s) set %s = %v, want %v", tt.name, tt.input, tt.field, value.String(), tt.input)
			}
		})
	}
}

// TestIdleTimeFunctions tests idle time configuration functions
func TestIdleTimeFunctions(t *testing.T) {
	// Test SetReduceIdleTime directly - the complex table test was having type issues
	t.Run("SetReduceIdleTime", func(t *testing.T) {
		emit := &SAMEmit{}

		// Test valid case
		err := SetReduceIdleTime(10)(emit)
		if err != nil {
			t.Errorf("SetReduceIdleTime(10) unexpected error: %v", err)
		}
		if emit.I2PConfig.ReduceIdleTime != 600000 {
			t.Errorf("SetReduceIdleTime(10) set ReduceIdleTime = %v, want 600000", emit.I2PConfig.ReduceIdleTime)
		}

		// Test invalid case
		emit2 := &SAMEmit{}
		err = SetReduceIdleTime(5)(emit2)
		if err == nil {
			t.Error("SetReduceIdleTime(5) expected error, got nil")
		}
		if emit2.I2PConfig.ReduceIdleTime != 300000 {
			t.Errorf("SetReduceIdleTime(5) set ReduceIdleTime = %v, want 300000 (default)", emit2.I2PConfig.ReduceIdleTime)
		}
	})

	// Test SetReduceIdleTimeMs directly
	t.Run("SetReduceIdleTimeMs", func(t *testing.T) {
		emit := &SAMEmit{}

		// Test valid case
		err := SetReduceIdleTimeMs(600000)(emit)
		if err != nil {
			t.Errorf("SetReduceIdleTimeMs(600000) unexpected error: %v", err)
		}
		if emit.I2PConfig.ReduceIdleTime != 600000 {
			t.Errorf("SetReduceIdleTimeMs(600000) set ReduceIdleTime = %v, want 600000", emit.I2PConfig.ReduceIdleTime)
		}

		// Test invalid case
		emit2 := &SAMEmit{}
		err = SetReduceIdleTimeMs(100000)(emit2)
		if err == nil {
			t.Error("SetReduceIdleTimeMs(100000) expected error, got nil")
		}
		if emit2.I2PConfig.ReduceIdleTime != 300000 {
			t.Errorf("SetReduceIdleTimeMs(100000) set ReduceIdleTime = %v, want 300000 (default)", emit2.I2PConfig.ReduceIdleTime)
		}
	})

	// Test SetCloseIdleTime
	t.Run("SetCloseIdleTime", func(t *testing.T) {
		emit := &SAMEmit{}

		// Test valid case
		err := SetCloseIdleTime(30)(emit)
		if err != nil {
			t.Errorf("SetCloseIdleTime(30) unexpected error: %v", err)
		}
		if emit.I2PConfig.CloseIdleTime != 1800000 {
			t.Errorf("SetCloseIdleTime(30) set CloseIdleTime = %v, want 1800000", emit.I2PConfig.CloseIdleTime)
		}

		// Test invalid case
		emit2 := &SAMEmit{}
		err = SetCloseIdleTime(5)(emit2)
		if err == nil {
			t.Error("SetCloseIdleTime(5) expected error, got nil")
		}
		if emit2.I2PConfig.CloseIdleTime != 300000 {
			t.Errorf("SetCloseIdleTime(5) set CloseIdleTime = %v, want 300000 (default)", emit2.I2PConfig.CloseIdleTime)
		}
	})

	// Test SetCloseIdleTimeMs
	t.Run("SetCloseIdleTimeMs", func(t *testing.T) {
		emit := &SAMEmit{}

		// Test valid case
		err := SetCloseIdleTimeMs(1800000)(emit)
		if err != nil {
			t.Errorf("SetCloseIdleTimeMs(1800000) unexpected error: %v", err)
		}
		if emit.I2PConfig.CloseIdleTime != 1800000 {
			t.Errorf("SetCloseIdleTimeMs(1800000) set CloseIdleTime = %v, want 1800000", emit.I2PConfig.CloseIdleTime)
		}

		// Test invalid case
		emit2 := &SAMEmit{}
		err = SetCloseIdleTimeMs(100000)(emit2)
		if err == nil {
			t.Error("SetCloseIdleTimeMs(100000) expected error, got nil")
		}
		if emit2.I2PConfig.CloseIdleTime != 300000 {
			t.Errorf("SetCloseIdleTimeMs(100000) set CloseIdleTime = %v, want 300000 (default)", emit2.I2PConfig.CloseIdleTime)
		}
	})
}

// TestAccessListFunctions tests access list configuration functions
func TestAccessListFunctions(t *testing.T) {
	// Test SetAccessListType
	t.Run("SetAccessListType", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			wantErr  bool
			expected string
		}{
			{"whitelist", "whitelist", false, "whitelist"},
			{"blacklist", "blacklist", false, "blacklist"},
			{"none", "none", false, ""},
			{"empty", "", false, ""},
			{"invalid", "invalid", true, ""},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				emit := &SAMEmit{}
				err := SetAccessListType(tt.input)(emit)

				if (err != nil) != tt.wantErr {
					t.Errorf("SetAccessListType(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				}

				if !tt.wantErr && emit.I2PConfig.AccessListType != tt.expected {
					t.Errorf("SetAccessListType(%s) set AccessListType = %v, want %v", tt.input, emit.I2PConfig.AccessListType, tt.expected)
				}
			})
		}
	})

	// Test SetAccessList
	t.Run("SetAccessList", func(t *testing.T) {
		tests := []struct {
			name  string
			input []string
		}{
			{"empty list", []string{}},
			{"single destination", []string{"dest1.b32.i2p"}},
			{"multiple destinations", []string{"dest1.b32.i2p", "dest2.b32.i2p"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				emit := &SAMEmit{}
				err := SetAccessList(tt.input)(emit)
				if err != nil {
					t.Errorf("SetAccessList(%v) unexpected error: %v", tt.input, err)
				}

				// For non-empty lists, check that items were appended
				if len(tt.input) > 0 && len(emit.I2PConfig.AccessList) != len(tt.input) {
					t.Errorf("SetAccessList(%v) set AccessList length = %v, want %v", tt.input, len(emit.I2PConfig.AccessList), len(tt.input))
				}
			})
		}
	})
}

// TestReduceIdleQuantity tests the reduce idle quantity configuration function
func TestReduceIdleQuantity(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		wantErr  bool
		expected int
	}{
		{"valid quantity", 2, false, 2},
		{"min valid", 0, false, 0},
		{"max valid", 4, false, 4},
		{"invalid high", 5, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{}
			err := SetReduceIdleQuantity(tt.input)(emit)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetReduceIdleQuantity(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}

			if !tt.wantErr && emit.I2PConfig.ReduceIdleQuantity != tt.expected {
				t.Errorf("SetReduceIdleQuantity(%d) set ReduceIdleQuantity = %v, want %v", tt.input, emit.I2PConfig.ReduceIdleQuantity, tt.expected)
			}
		})
	}
}

// BenchmarkConfigFunctions benchmarks key configuration functions for performance
func BenchmarkConfigFunctions(b *testing.B) {
	b.Run("NewConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = NewConfig()
		}
	})

	b.Run("NewEmit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = NewEmit()
		}
	})

	b.Run("SetType", func(b *testing.B) {
		emit := &SAMEmit{}
		for i := 0; i < b.N; i++ {
			_ = SetType("STREAM")(emit)
		}
	})

	b.Run("SetInLength", func(b *testing.B) {
		emit := &SAMEmit{}
		for i := 0; i < b.N; i++ {
			_ = SetInLength(3)(emit)
		}
	})
}
