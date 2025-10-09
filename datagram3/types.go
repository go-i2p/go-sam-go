package datagram3

import (
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
	"github.com/samber/oops"
)

// Datagram3Session represents a repliable but hash-based datagram3 session.
//
// DATAGRAM3 provides UDP-like messaging with hash-based source identification instead of
// full with full destinations destinations. This reduces overhead at the cost of full destination verification.
// Received datagrams contain a 32-byte hash that requires NAMING LOOKUP to resolve for replies.
//
// Key differences from DATAGRAM/DATAGRAM2:
//   - Repliable: Can reply to sender (like DATAGRAM/DATAGRAM2)
//   - Unwith full destinations: Source uses hash-based identification (unlike DATAGRAM/DATAGRAM2)
//   - Hash-based source: 32-byte hash instead of full destination
//   - Lower overhead: Hash-based identification required
//   - Reply overhead: Requires NAMING LOOKUP to resolve hash
//
// The session manages I2P tunnels and provides methods for creating readers and writers.
// For SAMv3 mode, it uses UDP forwarding where datagrams are received via a local UDP socket
// that the SAM bridge forwards to. The session maintains a hash resolver cache to avoid
// repeated NAMING LOOKUP operations when replying to the same source.
//
// I2P Timing Considerations:
//   - Session creation: 2-5 minutes for tunnel establishment
//   - Message delivery: Variable latency (network-dependent)
//   - Hash resolution: Additional network round-trip for NAMING LOOKUP
//   - Use generous timeouts and retry logic with exponential backoff
//
// Example usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//	session, err := NewDatagram3Session(sam, "my-session", keys, options)
//	reader := session.NewReader()
//	dg, err := reader.ReceiveDatagram()
//	if err := dg.ResolveSource(session); err != nil {
//	    log.Fatal(err)
//	}
//	session.NewWriter().SendDatagram(reply, dg.Source)
type Datagram3Session struct {
	*common.BaseSession
	sam        *common.SAM
	options    []string
	mu         sync.RWMutex
	closed     bool
	udpConn    *net.UDPConn  // UDP connection for receiving forwarded datagrams (SAMv3 mode)
	udpEnabled bool          // Whether UDP forwarding is enabled (always true for SAMv3)
	resolver   *HashResolver // Cache for hash-to-destination lookups
}

// Datagram3Reader handles incoming hash-based datagram3 reception from I2P.
//
// The reader provides asynchronous datagram reception through buffered channels, allowing
// applications to receive datagrams without blocking. It manages its own goroutine for
// continuous message processing and provides thread-safe access to received datagrams.
//
// Unlike DATAGRAM/DATAGRAM2, sources are represented as 32-byte hashes rather than full
// destinations. Applications must call ResolveSource() on received datagrams to obtain
// the full destination for replies. The session's resolver cache minimizes lookup overhead.
//
// Example usage:
//
//	reader := session.NewReader()
//	for {
//	    datagram, err := reader.ReceiveDatagram()
//	    if err != nil {
//	        // Handle error
//	    }
//	    // Verify using application-layer authentication before trusting
//	    if err := datagram.ResolveSource(session); err != nil {
//	        // Handle resolution error
//	    }
//	    // Now datagram.Source contains full destination for reply
//	}
type Datagram3Reader struct {
	session     *Datagram3Session
	recvChan    chan *Datagram3
	errorChan   chan error
	closeChan   chan struct{}
	doneChan    chan struct{}
	closed      bool
	loopStarted bool
	mu          sync.RWMutex
	closeOnce   sync.Once
}

// Datagram3Writer handles outgoing datagram3 transmission to I2P destinations.
// It provides methods for sending datagrams with configurable timeouts and handles
// the underlying SAM protocol communication for message delivery. The writer supports
// method chaining for configuration and provides error handling for send operations.
//
// Maximum datagram size is 31744 bytes total (including headers), with 11 KB recommended
// for best reliability. Destinations can be specified as full base64 destinations,
// hostnames (.i2p), or b32 addresses.
//
// Example usage:
//
//	writer := session.NewWriter().SetTimeout(30*time.Second)
//	err := writer.SendDatagram(data, destination)
type Datagram3Writer struct {
	session *Datagram3Session
	timeout time.Duration
}

// Datagram3 represents an I2P datagram3 message with source.
//
// This structure encapsulates the payload data along with the source hash
// and optional resolved destination. The SourceHash is always present (32 bytes), while
// Source is only populated after calling ResolveSource() to perform NAMING LOOKUP.
//
// Fields:
//   - Data: Raw datagram payload (up to ~31KB)
//   - SourceHash: 32-byte hash of sender (hash-based!)
//   - Source: Resolved full destination (nil until ResolveSource() called)
//   - Local: Local destination (this session)
//
// Example usage:
//
//	// Received datagram has only hash, not full source
//	log.Warn("Received from hash:", hex.EncodeToString(dg.SourceHash))
//
//	// Resolve hash to full destination for reply
//	if err := dg.ResolveSource(session); err != nil {
//	    return err
//	}
//
//	// Now can reply using resolved source (resolved!)
//	writer.SendDatagram(reply, dg.Source)
type Datagram3 struct {
	Data       []byte          // Raw datagram payload (up to ~31KB)
	SourceHash []byte          // 32-byte hash (hash-based!)
	Source     i2pkeys.I2PAddr // Resolved destination (nil until ResolveSource)
	Local      i2pkeys.I2PAddr // Local destination (this session)
}

// ResolveSource resolves the source hash to a full I2P destination for replying.
// This performs a NAMING LOOKUP to convert the 32-byte hash into a full destination
// address. The operation is cached in the session's resolver to avoid repeated lookups.
//
// Process:
//  1. Check if already resolved (Source not nil)
//  2. Validate SourceHash is 32 bytes
//  3. Convert hash to b32.i2p address (base32 encoding)
//  4. Perform NAMING LOOKUP via SAM bridge
//  5. Cache result in session resolver
//  6. Populate Source field with full destination
//
// This is an expensive operation (network round-trip) so results are cached.
// Applications replying to the same source repeatedly benefit from caching.
//
// Example usage:
//
//	if err := datagram.ResolveSource(session); err != nil {
//	    log.Error("Failed to resolve source:", err)
//	    return err
//	}
//	// datagram.Source now contains full destination
func (d *Datagram3) ResolveSource(session *Datagram3Session) error {
	// Validate input
	if session == nil {
		return oops.Errorf("session cannot be nil")
	}
	if len(d.SourceHash) == 0 {
		return oops.Errorf("no source hash available")
	}

	// Check if already resolved (I2PAddr is a string type, empty = not resolved)
	if d.Source != "" {
		return nil // Already resolved
	}

	// Resolve via session resolver (uses cache)
	dest, err := session.resolver.ResolveHash(d.SourceHash)
	if err != nil {
		return err
	}

	d.Source = dest
	return nil
}

// GetSourceB32 returns the b32.i2p address for the source hash without full resolution.
// This converts the 32-byte hash to a base32-encoded .b32.i2p address string without
// performing NAMING LOOKUP. This is faster than full resolution and sufficient for
// display, logging, or caching purposes.
//
// Returns empty string if SourceHash is invalid (not 32 bytes).
//
// Example usage:
//
//	b32Addr := datagram.GetSourceB32()
//	log.Info("Received from (unverified):", b32Addr)
func (d *Datagram3) GetSourceB32() string {
	if len(d.SourceHash) != 32 {
		return ""
	}

	return hashToB32Address(d.SourceHash)
}

// Datagram3Addr implements net.Addr interface for I2P datagram3 addresses.
//
// This type provides standard Go networking address representation for I2P destinations,
// allowing seamless integration with existing Go networking code that expects net.Addr.
// The address can wrap either a full I2P destination or just a hash from reception.
//
// Example usage:
//
//	addr := &Datagram3Addr{addr: destination, hash: sourceHash}
//	fmt.Println(addr.Network(), addr.String())
type Datagram3Addr struct {
	addr i2pkeys.I2PAddr
	hash []byte // Original 32-byte hash if from reception (hash-based!)
}

// Network returns the network type for I2P datagram3 addresses.
// This implements the net.Addr interface by returning "datagram3" as the network type.
func (a *Datagram3Addr) Network() string {
	return "datagram3"
}

// String returns the string representation of the I2P address.
// This implements the net.Addr interface. If a full address is available, returns base32
// representation. If only hash is available, returns the b32.i2p derived address.
func (a *Datagram3Addr) String() string {
	// Return full address if available (I2PAddr is a string type, empty = not available)
	if a.addr != "" {
		return a.addr.Base32()
	}
	// Fall back to hash-derived b32 address if only hash available
	if len(a.hash) == 32 {
		return hashToB32Address(a.hash)
	}
	return ""
}

// Datagram3Conn implements net.PacketConn interface for I2P datagram3 communication.
//
// This type provides compatibility with standard Go networking patterns by wrapping
// datagram3 session functionality in a familiar PacketConn interface. It manages
// internal readers and writers while providing standard connection operations.
//
// The connection provides thread-safe concurrent access to I2P datagram3 operations
// and properly handles cleanup on close. Unlike DATAGRAM/DATAGRAM2, sources are
// hash-based and not cryptographically verified.
//
// Example usage:
//
//	conn := session.PacketConn()
//	n, addr, err := conn.ReadFrom(buffer)
//	// addr represents source!
//	n, err = conn.WriteTo(data, destination)
type Datagram3Conn struct {
	session    *Datagram3Session
	reader     *Datagram3Reader
	writer     *Datagram3Writer
	remoteAddr *i2pkeys.I2PAddr
	mu         sync.RWMutex
	closed     bool
	cleanup    runtime.Cleanup
}
