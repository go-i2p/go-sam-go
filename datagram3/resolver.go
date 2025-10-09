package datagram3

import (
	"encoding/base32"
	"strings"
	"sync"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
)

// HashResolver provides caching for hash-to-destination lookups via NAMING LOOKUP.
// This prevents repeated network queries for the same hash, which is critical for
// DATAGRAM3 performance since every received datagram contains only a hash.
//
//
// The resolver maintains an in-memory cache mapping b32.i2p addresses to full I2P
// destinations. This cache is thread-safe using RWMutex and grows unbounded (applications
// should monitor memory usage for long-running sessions receiving from many sources).
//
// Hash Resolution Process:
//  1. Convert 32-byte hash to base32 (52 characters)
//  2. Append ".b32.i2p" suffix
//  3. Check cache for existing entry
//  4. If not cached, perform NAMING LOOKUP via SAM bridge
//  5. Cache successful result
//  6. Return full I2P destination
//
// Example usage:
//
//	resolver := NewHashResolver(sam)
//	dest, err := resolver.ResolveHash(hashBytes)
//	if err != nil {
//	    log.Error("Resolution failed:", err)
//	}
type HashResolver struct {
	sam   *common.SAM
	cache map[string]i2pkeys.I2PAddr // map[b32_address] -> full destination
	mu    sync.RWMutex
}

// NewHashResolver creates a new hash resolver with empty cache.
// The resolver uses the provided SAM connection for NAMING LOOKUP operations
// when cache misses occur.
//
// Example usage:
//
//	resolver := NewHashResolver(sam)
func NewHashResolver(sam *common.SAM) *HashResolver {
	return &HashResolver{
		sam:   sam,
		cache: make(map[string]i2pkeys.I2PAddr),
	}
}

// ResolveHash converts a 32-byte hash to a full I2P destination using NAMING LOOKUP.
//
//
// Process:
//  1. Validate hash is exactly 32 bytes
//  2. Convert to b32.i2p address (base32 encoding + suffix)
//  3. Check cache for existing result
//  4. If cached, return immediately (fast path)
//  5. If not cached, perform NAMING LOOKUP (slow path, network I/O)
//  6. Cache successful result for future lookups
//  7. Return full destination
//
// This is an expensive operation on cache misses due to network round-trip to I2P router.
// Applications should minimize unnecessary resolutions by caching at application level
// or reusing the same session resolver.
//
// Error conditions:
//   - Invalid hash length (not 32 bytes)
//   - Base32 encoding failure (malformed hash)
//   - NAMING LOOKUP failure (hash not resolvable, network error, etc.)
//
// Example usage:
//
//	dest, err := resolver.ResolveHash(datagram.SourceHash)
//	if err != nil {
//	    log.Error("Failed to resolve hash:", err)
//	    return err
//	}
//	// dest contains full I2P destination (resolved!)
//	writer.SendDatagram(reply, dest)
func (r *HashResolver) ResolveHash(hash []byte) (i2pkeys.I2PAddr, error) {
	// Validate hash length (CRITICAL: must be exactly 32 bytes)
	if len(hash) != 32 {
		return "", oops.Errorf("invalid hash length: %d (expected 32)", len(hash))
	}

	// Validate SAM connection is available
	if r.sam == nil {
		return "", oops.Errorf("SAM connection not available for hash resolution")
	}

	// Convert hash to b32.i2p address
	b32Addr := hashToB32Address(hash)

	// Check cache (read lock for concurrent access)
	r.mu.RLock()
	if dest, ok := r.cache[b32Addr]; ok {
		r.mu.RUnlock()
		log.WithField("b32", b32Addr).Debug("Hash resolved from cache")
		return dest, nil
	}
	r.mu.RUnlock()

	// Cache miss - perform NAMING LOOKUP (expensive network operation)
	log.WithField("b32", b32Addr).Debug("Cache miss - performing NAMING LOOKUP")
	dest, err := r.sam.Lookup(b32Addr)
	if err != nil {
		return "", oops.Errorf("NAMING LOOKUP failed for %s: %w", b32Addr, err)
	}

	// Cache successful result (write lock for exclusive access)
	r.mu.Lock()
	r.cache[b32Addr] = dest
	cacheSize := len(r.cache)
	r.mu.Unlock()

	log.WithField("b32", b32Addr).WithField("cache_size", cacheSize).Debug("Hash resolved and cached")
	return dest, nil
}

// GetCached returns cached destination without performing lookup.
// This allows checking if a hash has been previously resolved without triggering
// a potentially expensive NAMING LOOKUP operation.
//
// Returns:
//   - destination: Full I2P destination if cached
//   - found: true if entry exists in cache, false otherwise
//
// This method is useful for applications that want to avoid network I/O and only
// use already-resolved destinations. It's also useful for testing cache behavior.
//
// Example usage:
//
//	if dest, ok := resolver.GetCached(hash); ok {
//	    // Use cached destination without network lookup
//	    writer.SendDatagram(reply, dest)
//	} else {
//	    // Hash not yet resolved - decide whether to resolve now
//	    log.Info("Hash not in cache, resolution required for reply")
//	}
func (r *HashResolver) GetCached(hash []byte) (i2pkeys.I2PAddr, bool) {
	if len(hash) != 32 {
		return "", false
	}

	b32Addr := hashToB32Address(hash)

	r.mu.RLock()
	defer r.mu.RUnlock()

	dest, ok := r.cache[b32Addr]
	return dest, ok
}

// Clear removes all cached entries.
// This is useful for testing, memory management in long-running sessions, or when
// you want to force fresh NAMING LOOKUP operations.
//
//
// Applications with memory constraints may want to implement periodic cache clearing
// or LRU eviction policies on top of this basic cache.
//
// Example usage:
//
//	// Clear cache after processing batch
//	resolver.Clear()
//
//	// Or clear periodically
//	ticker := time.NewTicker(1 * time.Hour)
//	go func() {
//	    for range ticker.C {
//	        resolver.Clear()
//	    }
//	}()
func (r *HashResolver) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldSize := len(r.cache)
	r.cache = make(map[string]i2pkeys.I2PAddr)
	log.WithField("old_size", oldSize).Debug("Cache cleared")
}

// CacheSize returns the current number of cached entries.
// This is useful for monitoring memory usage and cache effectiveness.
//
// Example usage:
//
//	size := resolver.CacheSize()
//	log.Info("Cache contains", size, "entries")
func (r *HashResolver) CacheSize() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.cache)
}

// hashToB32Address converts a 32-byte hash to a b32.i2p address string.
// This performs base32 encoding (RFC 4648) and appends the ".b32.i2p" suffix.
//
// Format: <52-char-lowercase-base32>.b32.i2p
//
// The resulting address can be used for NAMING LOOKUP to resolve the full destination.
// This is a pure utility function with no network I/O.
//
// Example:
//
//	hash := []byte{0x01, 0x02, ...} // 32 bytes
//	addr := hashToB32Address(hash)
//	// addr = "aebagbafaydqqcikbmgq...xyz.b32.i2p"
func hashToB32Address(hash []byte) string {
	if len(hash) != 32 {
		return ""
	}

	// Base32 encode the hash (RFC 4648 standard encoding)
	// This produces 52 characters for 32 bytes of input
	b32 := base32.StdEncoding.EncodeToString(hash)

	// Convert to lowercase (I2P convention for b32 addresses)
	b32 = strings.ToLower(b32)

	// Remove padding if present (= characters)
	b32 = strings.TrimRight(b32, "=")

	// Append .b32.i2p suffix (I2P naming convention)
	return b32 + ".b32.i2p"
}
