package datagram3

import (
	"net"
	"testing"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// Compile-time interface checks to ensure types implement expected interfaces
var (
	_ common.Session = &Datagram3Session{}
	_ net.PacketConn = &Datagram3Conn{}
	_ net.Addr       = &Datagram3Addr{}
)

// TestDatagram3AddrImplementsNetAddr verifies Datagram3Addr properly implements net.Addr
func TestDatagram3AddrImplementsNetAddr(t *testing.T) {
	// Test with nil/empty address
	addr1 := &Datagram3Addr{}
	if addr1.Network() != "datagram3" {
		t.Errorf("Expected network 'datagram3', got %q", addr1.Network())
	}

	// Test with hash
	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = byte(i)
	}
	addr2 := &Datagram3Addr{hash: hash}
	if addr2.Network() != "datagram3" {
		t.Errorf("Expected network 'datagram3', got %q", addr2.Network())
	}

	// Test String() returns something
	if addr2.String() == "" {
		t.Error("Expected non-empty string representation")
	}
}

// TestDatagram3GetSourceB32 tests the hash-to-b32 conversion without full resolution
func TestDatagram3GetSourceB32(t *testing.T) {
	tests := []struct {
		name     string
		hash     []byte
		wantLen  int
		wantErr  bool
		wantZero bool
	}{
		{
			name:     "valid 32-byte hash",
			hash:     make([]byte, 32),
			wantLen:  60, // 52 chars base32 + ".b32.i2p" (8 chars)
			wantErr:  false,
			wantZero: false,
		},
		{
			name:     "nil hash",
			hash:     nil,
			wantLen:  0,
			wantErr:  false,
			wantZero: true,
		},
		{
			name:     "empty hash",
			hash:     []byte{},
			wantLen:  0,
			wantErr:  false,
			wantZero: true,
		},
		{
			name:     "invalid hash length (31 bytes)",
			hash:     make([]byte, 31),
			wantLen:  0,
			wantErr:  false,
			wantZero: true,
		},
		{
			name:     "invalid hash length (33 bytes)",
			hash:     make([]byte, 33),
			wantLen:  0,
			wantErr:  false,
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dg := &Datagram3{
				SourceHash: tt.hash,
			}

			result := dg.GetSourceB32()

			if tt.wantZero {
				if result != "" {
					t.Errorf("Expected empty string for invalid hash, got %q", result)
				}
				return
			}

			if len(result) != tt.wantLen {
				t.Errorf("Expected b32 address length %d, got %d", tt.wantLen, len(result))
			}

			// Verify it ends with .b32.i2p
			if len(result) > 0 && result[len(result)-8:] != ".b32.i2p" {
				t.Errorf("Expected address to end with '.b32.i2p', got %q", result)
			}
		})
	}
}

// TestDatagram3ResolveSourceValidation tests validation before resolution
func TestDatagram3ResolveSourceValidation(t *testing.T) {
	// Create a dummy session for testing
	dummySession := &Datagram3Session{
		resolver: &HashResolver{
			sam:   nil,
			cache: make(map[string]i2pkeys.I2PAddr),
		},
	}

	// Test with nil session
	dg0 := &Datagram3{SourceHash: make([]byte, 32)}
	err := dg0.ResolveSource(nil)
	if err == nil {
		t.Error("Expected error when session is nil")
	}

	// Test with no hash
	dg1 := &Datagram3{}
	err = dg1.ResolveSource(dummySession)
	if err == nil {
		t.Error("Expected error when resolving without hash")
	}

	// Test with already resolved source
	dg2 := &Datagram3{
		SourceHash: make([]byte, 32),
		Source:     i2pkeys.I2PAddr("test"),
	}
	err = dg2.ResolveSource(dummySession)
	if err != nil {
		t.Errorf("Expected no error for already resolved source, got %v", err)
	}
}

// TestHashResolverCacheOperations tests the cache management without I2P
func TestHashResolverCacheOperations(t *testing.T) {
	// Create resolver with nil SAM (cache-only operations)
	resolver := &HashResolver{
		sam:   nil,
		cache: make(map[string]i2pkeys.I2PAddr),
	}

	// Test CacheSize on empty cache
	if size := resolver.CacheSize(); size != 0 {
		t.Errorf("Expected cache size 0, got %d", size)
	}

	// Add entries to cache manually for testing
	hash1 := make([]byte, 32)
	for i := range hash1 {
		hash1[i] = 1
	}
	b32_1 := hashToB32Address(hash1)
	resolver.cache[b32_1] = i2pkeys.I2PAddr("destination1")

	hash2 := make([]byte, 32)
	for i := range hash2 {
		hash2[i] = 2
	}
	b32_2 := hashToB32Address(hash2)
	resolver.cache[b32_2] = i2pkeys.I2PAddr("destination2")

	// Test CacheSize
	if size := resolver.CacheSize(); size != 2 {
		t.Errorf("Expected cache size 2, got %d", size)
	}

	// Test GetCached for hit
	dest, ok := resolver.GetCached(hash1)
	if !ok {
		t.Error("Expected cache hit for hash1")
	}
	if dest != "destination1" {
		t.Errorf("Expected destination1, got %v", dest)
	}

	// Test GetCached for miss
	hash3 := make([]byte, 32)
	for i := range hash3 {
		hash3[i] = 3
	}
	_, ok = resolver.GetCached(hash3)
	if ok {
		t.Error("Expected cache miss for hash3")
	}

	// Test Clear
	resolver.Clear()
	if size := resolver.CacheSize(); size != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", size)
	}
}

// TestHashToB32AddressConversion tests the hash-to-b32 conversion logic
func TestHashToB32AddressConversion(t *testing.T) {
	tests := []struct {
		name    string
		hash    []byte
		wantLen int
	}{
		{
			name:    "all zeros",
			hash:    make([]byte, 32),
			wantLen: 60, // 52 + 8 (.b32.i2p)
		},
		{
			name: "all ones",
			hash: func() []byte {
				h := make([]byte, 32)
				for i := range h {
					h[i] = 0xFF
				}
				return h
			}(),
			wantLen: 60,
		},
		{
			name: "sequential bytes",
			hash: func() []byte {
				h := make([]byte, 32)
				for i := range h {
					h[i] = byte(i)
				}
				return h
			}(),
			wantLen: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b32 := hashToB32Address(tt.hash)

			if len(b32) != tt.wantLen {
				t.Errorf("Expected b32 address length %d, got %d", tt.wantLen, len(b32))
			}

			if b32[len(b32)-8:] != ".b32.i2p" {
				t.Errorf("Expected address to end with '.b32.i2p', got %q", b32)
			}

			// Verify it's lowercase (base32 standard)
			for i, c := range b32[:len(b32)-8] { // Don't check the suffix
				if c >= 'A' && c <= 'Z' {
					t.Errorf("Expected lowercase base32, found uppercase at position %d: %c", i, c)
				}
			}
		})
	}
}

// TestResolveHashInvalidLength tests that invalid hash lengths are rejected
func TestResolveHashInvalidLength(t *testing.T) {
	resolver := &HashResolver{
		sam:   nil, // Will fail before SAM access
		cache: make(map[string]i2pkeys.I2PAddr),
	}

	tests := []struct {
		name     string
		hashLen  int
		wantFail bool
	}{
		{"0 bytes", 0, true},
		{"1 byte", 1, true},
		{"16 bytes", 16, true},
		{"31 bytes", 31, true},
		{"33 bytes", 33, true},
		{"64 bytes", 64, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := make([]byte, tt.hashLen)
			_, err := resolver.ResolveHash(hash)

			if tt.wantFail {
				if err == nil {
					t.Errorf("Expected error for hash length %d, got nil", tt.hashLen)
				}
			}
		})
	}

	// Test valid hash length (32 bytes) separately since it will try SAM lookup
	t.Run("32 bytes valid length", func(t *testing.T) {
		hash := make([]byte, 32)
		_, err := resolver.ResolveHash(hash)
		// Will fail due to nil SAM, but shouldn't panic or fail length check
		if err == nil {
			t.Error("Expected error due to nil SAM, got nil")
		}
		// Just verify it doesn't panic and returns an error (SAM or lookup related)
	})
}
