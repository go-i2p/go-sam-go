package common

import (
	"strings"
	"testing"
)

func TestSetSAMAddress_Cases(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		wantHost string
		wantPort int
	}{
		{
			name:     "empty address uses defaults",
			addr:     "",
			wantHost: "127.0.0.1",
			wantPort: 7656,
		},
		{
			name:     "valid host:port",
			addr:     "192.168.1.1:7000",
			wantHost: "192.168.1.1",
			wantPort: 7000,
		},
		{
			name:     "invalid port uses default",
			addr:     "localhost:99999",
			wantHost: "localhost",
			wantPort: 7656, // Default port
		},
		{
			name:     "just IP address",
			addr:     "192.168.1.1",
			wantHost: "192.168.1.1",
			wantPort: 7656,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &I2PConfig{}
			cfg.SetSAMAddress(tt.addr)

			if cfg.SamHost != tt.wantHost {
				t.Errorf("SetSAMAddress() host = %v, want %v", cfg.SamHost, tt.wantHost)
			}
			if cfg.SamPort != tt.wantPort {
				t.Errorf("SetSAMAddress() port = %v, want %v", cfg.SamPort, tt.wantPort)
			}
		})
	}
}

func TestID_Generation(t *testing.T) {
	cfg := &I2PConfig{}

	// Test when TunName is empty
	id1 := cfg.ID()
	if len(cfg.TunName) != 12 {
		t.Errorf("ID() generated name length = %v, want 12", len(cfg.TunName))
	}

	// Verify format
	if id1[:3] != "ID=" {
		t.Errorf("ID() format incorrect, got %v, want prefix 'ID='", id1)
	}

	// Test with preset TunName
	cfg.TunName = "testtunnel"
	id2 := cfg.ID()
	if id2 != "ID=testtunnel" {
		t.Errorf("ID() = %v, want ID=testtunnel", id2)
	}
}

func TestSam_AddressFormatting(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{
			name: "default values",
			host: "",
			port: 0,
			want: "127.0.0.1:7656",
		},
		{
			name: "custom host and port",
			host: "localhost",
			port: 7000,
			want: "localhost:7000",
		},
		{
			name: "only custom host",
			host: "192.168.1.1",
			port: 0,
			want: "192.168.1.1:7656",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &I2PConfig{
				SamHost: tt.host,
				SamPort: tt.port,
			}
			got := cfg.Sam()
			if got != tt.want {
				t.Errorf("Sam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLeaseSetSettings_Formatting(t *testing.T) {
	tests := []struct {
		name        string
		cfg         I2PConfig
		wantKey     string
		wantPrivKey string
		wantSignKey string
	}{
		{
			name:        "empty settings",
			cfg:         I2PConfig{},
			wantKey:     "",
			wantPrivKey: "",
			wantSignKey: "",
		},
		{
			name: "all settings populated",
			cfg: I2PConfig{
				LeaseSetKey:               "testkey",
				LeaseSetPrivateKey:        "privkey",
				LeaseSetPrivateSigningKey: "signkey",
			},
			wantKey:     " i2cp.leaseSetKey=testkey ",
			wantPrivKey: " i2cp.leaseSetPrivateKey=privkey ",
			wantSignKey: " i2cp.leaseSetPrivateSigningKey=signkey ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, priv, sign := tt.cfg.LeaseSetSettings()
			if key != tt.wantKey {
				t.Errorf("LeaseSetSettings() key = %v, want %v", key, tt.wantKey)
			}
			if priv != tt.wantPrivKey {
				t.Errorf("LeaseSetSettings() private key = %v, want %v", priv, tt.wantPrivKey)
			}
			if sign != tt.wantSignKey {
				t.Errorf("LeaseSetSettings() signing key = %v, want %v", sign, tt.wantSignKey)
			}
		})
	}
}

func TestTunnelConfigNoDuplicates(t *testing.T) {
	cfg := &I2PConfig{
		InLength:    3,
		OutLength:   3,
		InQuantity:  2,
		OutQuantity: 2,
	}

	config := cfg.TunnelConfig()
	params := strings.Split(config, " ")

	// Verify no duplicate parameters
	seen := make(map[string]bool)
	for _, param := range params {
		if seen[param] {
			t.Errorf("Duplicate parameter found: %s", param)
		}
		seen[param] = true
	}

	// Verify expected parameters present
	expectedParams := []string{
		"inbound.length=3",
		"outbound.length=3",
		"inbound.quantity=2",
		"outbound.quantity=2",
	}

	for _, expected := range expectedParams {
		if !strings.Contains(config, expected) {
			t.Errorf("Missing expected parameter: %s", expected)
		}
	}
}
