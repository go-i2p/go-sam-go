package common

import "testing"

func TestSetInQuantity(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"valid min", 1, false},
		{"valid max", 16, false},
		{"valid middle", 8, false},
		{"invalid zero", 0, true},
		{"invalid negative", -1, true},
		{"invalid too large", 17, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{I2PConfig: I2PConfig{}}
			err := SetInQuantity(tt.input)(emit)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("SetInQuantity() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			if !tt.wantErr && emit.I2PConfig.InQuantity != tt.input {
				t.Errorf("SetInQuantity() = %v, want %v", emit.I2PConfig.InQuantity, tt.input)
			}
		})
	}
}

func TestSetOutQuantity(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"valid min", 1, false},
		{"valid max", 16, false},
		{"valid middle", 8, false},
		{"invalid zero", 0, true},
		{"invalid negative", -1, true},
		{"invalid too large", 17, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emit := &SAMEmit{I2PConfig: I2PConfig{}}
			err := SetOutQuantity(tt.input)(emit)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("SetOutQuantity() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}

			if !tt.wantErr && emit.I2PConfig.OutQuantity != tt.input {
				t.Errorf("SetOutQuantity() = %v, want %v", emit.I2PConfig.OutQuantity, tt.input)
			}
		})
	}
}
