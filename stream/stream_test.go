package stream

import (
	"testing"
)

func TestNewStreamSession_Integration(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		options []string
		wantErr bool
	}{
		{
			name:    "basic session",
			id:      "test1",
			options: nil,
			wantErr: false,
		},
		{
			name:    "with options",
			id:      "test2",
			options: []string{"inbound.length=2", "outbound.length=2"},
			wantErr: false,
		},
		{
			name:    "invalid options",
			id:      "test3",
			options: []string{"invalid=true"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh SAM connection for each test
			sam, err := NewSAM("127.0.0.1:7656")
			if err != nil {
				t.Fatalf("NewSAM() error = %v", err)
			}

			// Generate keys through the SAM bridge
			keys, err := sam.NewKeys()
			if err != nil {
				t.Fatalf("NewKeys() error = %v", err)
			}

			session, err := sam.NewStreamSession(tt.id, keys, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStreamSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				session.Close()
			}
			sam.Close()
		})
	}
}
