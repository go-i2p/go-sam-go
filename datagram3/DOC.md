# datagram3
--
    import "github.com/go-i2p/go-sam-go/datagram3"

Package datagram3 provides repliable datagram sessions with hash-based source
identification for I2P.

DATAGRAM3 sessions provide repliable UDP-like messaging with hash-based source
identification instead of full destinations.

Key features:

    - Repliable (can send replies to sender)
    - Hash-based source identification (32-byte hash)
    - Requires NAMING LOOKUP for replies
    - UDP-like messaging (unreliable, unordered)
    - Maximum 31744 bytes per datagram (11 KB recommended)

Session creation requires 2-5 minutes for I2P tunnel establishment. Use generous
timeouts and exponential backoff retry logic. Hash resolution uses automatic
caching to minimize NAMING LOOKUP overhead.

Basic usage:

    sam, err := common.NewSAM("127.0.0.1:7656")
    session, err := datagram3.NewDatagram3Session(sam, "my-session", keys, []string{"inbound.length=1"})
    defer session.Close()
    dg, err := session.NewReader().ReceiveDatagram()
    if err := dg.ResolveSource(session); err != nil { log.Error(err) }
    session.NewWriter().SendDatagram([]byte("reply"), dg.Source)

See also: Package datagram, datagram2, stream, raw, primary.

## Usage

#### type Datagram3

```go
type Datagram3 struct {
	Data       []byte          // Raw datagram payload (up to ~31KB)
	SourceHash []byte          // 32-byte hash (hash-based!)
	Source     i2pkeys.I2PAddr // Resolved destination (nil until ResolveSource)
	Local      i2pkeys.I2PAddr // Local destination (this session)
}
```

Datagram3 represents an I2P datagram3 message with source.

This structure encapsulates the payload data along with the source hash and
optional resolved destination. The SourceHash is always present (32 bytes),
while Source is only populated after calling ResolveSource() to perform NAMING
LOOKUP.

Fields:

    - Data: Raw datagram payload (up to ~31KB)
    - SourceHash: 32-byte hash of sender (hash-based!)
    - Source: Resolved full destination (nil until ResolveSource() called)
    - Local: Local destination (this session)

Example usage:

    // Received datagram has only hash, not full source
    log.Warn("Received from hash:", hex.EncodeToString(dg.SourceHash))

    // Resolve hash to full destination for reply
    if err := dg.ResolveSource(session); err != nil {
        return err
    }

    // Now can reply using resolved source (resolved!)
    writer.SendDatagram(reply, dg.Source)

#### func (*Datagram3) GetSourceB32

```go
func (d *Datagram3) GetSourceB32() string
```
GetSourceB32 returns the b32.i2p address for the source hash without full
resolution. This converts the 32-byte hash to a base32-encoded .b32.i2p address
string without performing NAMING LOOKUP. This is faster than full resolution and
sufficient for display, logging, or caching purposes.

Returns empty string if SourceHash is invalid (not 32 bytes).

Example usage:

    b32Addr := datagram.GetSourceB32()
    log.Info("Received from (unverified):", b32Addr)

#### func (*Datagram3) ResolveSource

```go
func (d *Datagram3) ResolveSource(session *Datagram3Session) error
```
ResolveSource resolves the source hash to a full I2P destination for replying.
This performs a NAMING LOOKUP to convert the 32-byte hash into a full
destination address. The operation is cached in the session's resolver to avoid
repeated lookups.

Process:

    1. Check if already resolved (Source not nil)
    2. Validate SourceHash is 32 bytes
    3. Convert hash to b32.i2p address (base32 encoding)
    4. Perform NAMING LOOKUP via SAM bridge
    5. Cache result in session resolver
    6. Populate Source field with full destination

This is an expensive operation (network round-trip) so results are cached.
Applications replying to the same source repeatedly benefit from caching.

Example usage:

    if err := datagram.ResolveSource(session); err != nil {
        log.Error("Failed to resolve source:", err)
        return err
    }
    // datagram.Source now contains full destination

#### type Datagram3Addr

```go
type Datagram3Addr struct {
}
```

Datagram3Addr implements net.Addr interface for I2P datagram3 addresses.

This type provides standard Go networking address representation for I2P
destinations, allowing seamless integration with existing Go networking code
that expects net.Addr. The address can wrap either a full I2P destination or
just a hash from reception.

Example usage:

    addr := &Datagram3Addr{addr: destination, hash: sourceHash}
    fmt.Println(addr.Network(), addr.String())

#### func (*Datagram3Addr) Network

```go
func (a *Datagram3Addr) Network() string
```
Network returns the network type for I2P datagram3 addresses. This implements
the net.Addr interface by returning "datagram3" as the network type.

#### func (*Datagram3Addr) String

```go
func (a *Datagram3Addr) String() string
```
String returns the string representation of the I2P address. This implements the
net.Addr interface. If a full address is available, returns base32
representation. If only hash is available, returns the b32.i2p derived address.

#### type Datagram3Conn

```go
type Datagram3Conn struct {
}
```

Datagram3Conn implements net.PacketConn interface for I2P datagram3
communication.

This type provides compatibility with standard Go networking patterns by
wrapping datagram3 session functionality in a familiar PacketConn interface. It
manages internal readers and writers while providing standard connection
operations.

The connection provides thread-safe concurrent access to I2P datagram3
operations and properly handles cleanup on close. Unlike DATAGRAM/DATAGRAM2,
sources are hash-based and not cryptographically verified.

Example usage:

    conn := session.PacketConn()
    n, addr, err := conn.ReadFrom(buffer)
    // addr represents source!
    n, err = conn.WriteTo(data, destination)

#### func (*Datagram3Conn) Close

```go
func (c *Datagram3Conn) Close() error
```
Close closes the datagram3 connection and releases associated resources. This
method implements the net.Conn interface. It closes the reader and writer but
does not close the underlying session, which may be shared by other connections.
Multiple calls to Close are safe and will return nil after the first call.

#### func (*Datagram3Conn) LocalAddr

```go
func (c *Datagram3Conn) LocalAddr() net.Addr
```
LocalAddr returns the local network address as a Datagram3Addr containing the
I2P destination address of this connection's session. This method implements the
net.Conn interface and provides access to the local I2P destination.

#### func (*Datagram3Conn) Read

```go
func (c *Datagram3Conn) Read(b []byte) (n int, err error)
```
Read implements net.Conn by wrapping ReadFrom for stream-like usage. It reads
data into the provided byte slice and returns the number of bytes read. When
reading, it also updates the remote address of the connection for subsequent
Write calls.

Note: This is not typical for datagrams which are connectionless, but provides
compatibility with the net.Conn interface.

#### func (*Datagram3Conn) ReadFrom

```go
func (c *Datagram3Conn) ReadFrom(p []byte) (n int, addr net.Addr, err error)
```
ReadFrom reads a datagram from the connection.

This method implements the net.PacketConn interface. It starts the receive loop
if not already started and blocks until a datagram is received. The data is
copied to the provided buffer p, and the source address is returned as a
Datagram3Addr.

The source address contains the 32-byte hash (not full destination).
Applications must resolve the hash via ResolveSource() to reply.

#### func (*Datagram3Conn) RemoteAddr

```go
func (c *Datagram3Conn) RemoteAddr() net.Addr
```
RemoteAddr returns the remote network address of the connection. This method
implements the net.Conn interface. For datagram3 connections, this returns the
address of the last peer that sent data (set by Read), or nil if no data has
been received yet.

#### func (*Datagram3Conn) SetDeadline

```go
func (c *Datagram3Conn) SetDeadline(t time.Time) error
```
SetDeadline sets both read and write deadlines for the connection. This method
implements the net.Conn interface by calling both SetReadDeadline and
SetWriteDeadline with the same time value. If either deadline cannot be set, the
first error encountered is returned.

#### func (*Datagram3Conn) SetReadDeadline

```go
func (c *Datagram3Conn) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets the deadline for future ReadFrom calls. This method
implements the net.Conn interface. For datagram3 connections, this is currently
a placeholder implementation that always returns nil. Timeout handling is
managed differently for datagram operations.

#### func (*Datagram3Conn) SetWriteDeadline

```go
func (c *Datagram3Conn) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets the deadline for future WriteTo calls. This method
implements the net.Conn interface. If the deadline is not zero, it calculates
the timeout duration and sets it on the writer for subsequent write operations.

#### func (*Datagram3Conn) Write

```go
func (c *Datagram3Conn) Write(b []byte) (n int, err error)
```
Write implements net.Conn by wrapping WriteTo for stream-like usage. It writes
data to the remote address set by the last Read operation and returns the number
of bytes written. If no remote address has been set, it returns an error. Note:
This is not typical for datagrams which are connectionless, but provides
compatibility with the net.Conn interface.

#### func (*Datagram3Conn) WriteTo

```go
func (c *Datagram3Conn) WriteTo(p []byte, addr net.Addr) (n int, err error)
```
WriteTo writes a datagram to the specified address. This method implements the
net.PacketConn interface. The address must be a Datagram3Addr or i2pkeys.I2PAddr
containing a valid I2P destination. The entire byte slice p is sent as a single
datagram message.

If the address is a Datagram3Addr with only a hash (not resolved), the hash will
be resolved automatically before sending.

#### type Datagram3Reader

```go
type Datagram3Reader struct {
}
```

Datagram3Reader handles incoming hash-based datagram3 reception from I2P.

The reader provides asynchronous datagram reception through buffered channels,
allowing applications to receive datagrams without blocking. It manages its own
goroutine for continuous message processing and provides thread-safe access to
received datagrams.

Unlike DATAGRAM/DATAGRAM2, sources are represented as 32-byte hashes rather than
full destinations. Applications must call ResolveSource() on received datagrams
to obtain the full destination for replies. The session's resolver cache
minimizes lookup overhead.

Example usage:

    reader := session.NewReader()
    for {
        datagram, err := reader.ReceiveDatagram()
        if err != nil {
            // Handle error
        }
        // Verify using application-layer authentication before trusting
        if err := datagram.ResolveSource(session); err != nil {
            // Handle resolution error
        }
        // Now datagram.Source contains full destination for reply
    }

#### func (*Datagram3Reader) Close

```go
func (r *Datagram3Reader) Close() error
```
Close closes the Datagram3Reader and stops its receive loop. This method safely
terminates the reader, cleans up all associated resources, and signals any
waiting goroutines to stop. It's safe to call multiple times and will not block
if the reader is already closed.

Example usage:

    defer reader.Close()

#### func (*Datagram3Reader) ReceiveDatagram

```go
func (r *Datagram3Reader) ReceiveDatagram() (*Datagram3, error)
```
ReceiveDatagram receives a single datagram from the I2P network.

This method blocks until a datagram is received or an error occurs, returning
the received datagram with its data and hash-based source. It handles concurrent
access safely and provides proper error handling for network issues.

Unlike DATAGRAM/DATAGRAM2, received datagrams contain only a 32-byte hash (not
full destination). Applications must call ResolveSource() to convert the hash to
a full destination for replies.

Example usage:

    datagram, err := reader.ReceiveDatagram()
    if err != nil {
        // Handle error
    }
    log.Info("Received from source:", hex.EncodeToString(datagram.SourceHash))
    if err := datagram.ResolveSource(session); err != nil {
        log.Error(err)
    }

#### type Datagram3Session

```go
type Datagram3Session struct {
	*common.BaseSession
}
```

Datagram3Session represents a repliable but hash-based datagram3 session.

DATAGRAM3 provides UDP-like messaging with hash-based source identification
instead of full with full destinations destinations. This reduces overhead at
the cost of full destination verification. Received datagrams contain a 32-byte
hash that requires NAMING LOOKUP to resolve for replies.

Key differences from DATAGRAM/DATAGRAM2:

    - Repliable: Can reply to sender (like DATAGRAM/DATAGRAM2)
    - Unwith full destinations: Source uses hash-based identification (unlike DATAGRAM/DATAGRAM2)
    - Hash-based source: 32-byte hash instead of full destination
    - Lower overhead: Hash-based identification required
    - Reply overhead: Requires NAMING LOOKUP to resolve hash

The session manages I2P tunnels and provides methods for creating readers and
writers. For SAMv3 mode, it uses UDP forwarding where datagrams are received via
a local UDP socket that the SAM bridge forwards to. The session maintains a hash
resolver cache to avoid repeated NAMING LOOKUP operations when replying to the
same source.

I2P Timing Considerations:

    - Session creation: 2-5 minutes for tunnel establishment
    - Message delivery: Variable latency (network-dependent)
    - Hash resolution: Additional network round-trip for NAMING LOOKUP
    - Use generous timeouts and retry logic with exponential backoff

Example usage:

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    session, err := NewDatagram3Session(sam, "my-session", keys, options)
    reader := session.NewReader()
    dg, err := reader.ReceiveDatagram()
    if err := dg.ResolveSource(session); err != nil {
        log.Fatal(err)
    }
    session.NewWriter().SendDatagram(reply, dg.Source)

#### func  NewDatagram3Session

```go
func NewDatagram3Session(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*Datagram3Session, error)
```
NewDatagram3Session creates a new datagram3 session with hash-based source
identification. It initializes the session with the provided SAM connection,
session ID, cryptographic keys, and configuration options. The session
automatically creates a UDP listener for receiving forwarded datagrams per SAMv3
requirements and initializes a hash resolver for source lookups. Note: DATAGRAM3
sources are not with full destinations; use datagram2 if authentication is
required. Example usage: session, err := NewDatagram3Session(sam, "my-session",
keys, []string{"inbound.length=1"})

#### func  NewDatagram3SessionFromSubsession

```go
func NewDatagram3SessionFromSubsession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, udpConn *net.UDPConn) (*Datagram3Session, error)
```
NewDatagram3SessionFromSubsession creates a Datagram3Session for a subsession
that has already been registered with a PRIMARY session using SESSION ADD. This
constructor skips the session creation step since the subsession is already
registered with the SAM bridge.

For PRIMARY datagram3 subsessions, UDP forwarding is mandatory (SAMv3
requirement). The UDP connection must be provided for proper datagram reception.
Note: Sources are not with full destinations; use NewDatagramSubSession if
authentication is required.

Example usage: sub, err := NewDatagram3SessionFromSubsession(sam, "sub1", keys,
options, udpConn)

#### func (*Datagram3Session) Addr

```go
func (s *Datagram3Session) Addr() i2pkeys.I2PAddr
```
Addr returns the local I2P address of this datagram3 session. This is the
destination address that other I2P nodes can use to send datagrams to this
session.

#### func (*Datagram3Session) Close

```go
func (s *Datagram3Session) Close() error
```
Close terminates the datagram3 session and cleans up all resources. This method
ensures proper cleanup of the UDP connection and I2P tunnels. After calling
Close(), the session cannot be reused. Example usage: defer session.Close()

#### func (*Datagram3Session) NewReader

```go
func (s *Datagram3Session) NewReader() *Datagram3Reader
```
NewReader creates a Datagram3Reader for receiving datagrams with hash-based
sources. This method initializes a new reader with buffered channels for
asynchronous datagram reception. The reader must be started manually with
receiveLoop() for continuous operation. Received datagrams contain 32-byte
hashes; call ResolveSource() to obtain full destinations for replies. Example
usage: reader := session.NewReader(); go reader.receiveLoop(); datagram, err :=
reader.ReceiveDatagram()

#### func (*Datagram3Session) NewWriter

```go
func (s *Datagram3Session) NewWriter() *Datagram3Writer
```
NewWriter creates a Datagram3Writer for sending datagrams to I2P destinations.
This method initializes a new writer with a default timeout of 30 seconds for
send operations. The timeout can be customized using the SetTimeout method on
the returned writer. Example usage: writer :=
session.NewWriter().SetTimeout(60*time.Second); err := writer.SendDatagram(data,
dest)

#### func (*Datagram3Session) PacketConn

```go
func (s *Datagram3Session) PacketConn() net.PacketConn
```
PacketConn returns a net.PacketConn interface for this datagram3 session. This
method provides compatibility with standard Go networking code by wrapping the
datagram3 session in a PacketConn interface. The returned connection manages its
own reader and writer and implements all standard net.PacketConn methods.

The connection is automatically cleaned up by a finalizer if Close() is not
called, but explicit Close() calls are strongly recommended to prevent resource
leaks.

Example usage:

    conn := session.PacketConn()
    defer conn.Close()

    // Receive source
    n, addr, err := conn.ReadFrom(buffer)

    // Send reply
    n, err = conn.WriteTo(reply, addr)

#### type Datagram3Writer

```go
type Datagram3Writer struct {
}
```

Datagram3Writer handles outgoing datagram3 transmission to I2P destinations. It
provides methods for sending datagrams with configurable timeouts and handles
the underlying SAM protocol communication for message delivery. The writer
supports method chaining for configuration and provides error handling for send
operations.

Maximum datagram size is 31744 bytes total (including headers), with 11 KB
recommended for best reliability. Destinations can be specified as full base64
destinations, hostnames (.i2p), or b32 addresses.

Example usage:

    writer := session.NewWriter().SetTimeout(30*time.Second)
    err := writer.SendDatagram(data, destination)

#### func (*Datagram3Writer) ReplyToDatagram

```go
func (w *Datagram3Writer) ReplyToDatagram(data []byte, original *Datagram3) error
```
ReplyToDatagram sends a reply to a received DATAGRAM3 message.

This automatically resolves the source hash if not already resolved, then sends
the reply. The source hash is resolved via NAMING LOOKUP and cached to avoid
repeated lookups.

Example usage:

    // Receive datagram
    dg, err := reader.ReceiveDatagram()
    if err != nil {
        return err
    }

    // Reply (automatically resolves hash)
    writer := session.NewWriter()
    err = writer.ReplyToDatagram([]byte("reply"), dg)

#### func (*Datagram3Writer) SendDatagram

```go
func (w *Datagram3Writer) SendDatagram(data []byte, dest i2pkeys.I2PAddr) error
```
SendDatagram sends a datagram to the specified I2P destination.

This method uses the SAMv3 UDP approach: sending via UDP socket to port 7655
with DATAGRAM3 format. The destination can be:

    - Full base64 destination (516+ chars)
    - Hostname (.i2p address)
    - B32 address (52 chars + .b32.i2p)
    - B32 address derived from received DATAGRAM3 hash (via ResolveSource)

Maximum datagram size is 31744 bytes total (including headers), with 11 KB
recommended for best reliability across the I2P network. It blocks until the
datagram is sent or an error occurs, respecting the configured timeout.

Example usage:

    // Send to full destination
    err := writer.SendDatagram([]byte("hello world"), destinationAddr)

    // Reply to received datagram (requires hash resolution)
    if err := receivedDatagram.ResolveSource(session); err != nil {
        return err
    }
    err := writer.SendDatagram([]byte("reply"), receivedDatagram.Source)

#### func (*Datagram3Writer) SetTimeout

```go
func (w *Datagram3Writer) SetTimeout(timeout time.Duration) *Datagram3Writer
```
SetTimeout sets the timeout for datagram3 write operations. This method
configures the maximum time to wait for datagram send operations to complete.
The timeout prevents indefinite blocking during network congestion or connection
issues. Returns the writer instance for method chaining convenience.

Example usage:

    writer.SetTimeout(30*time.Second).SendDatagram(data, destination)

#### type HashResolver

```go
type HashResolver struct {
}
```

HashResolver provides caching for hash-to-destination lookups via NAMING LOOKUP.
This prevents repeated network queries for the same hash, which is critical for
DATAGRAM3 performance since every received datagram contains only a hash.

The resolver maintains an in-memory cache mapping b32.i2p addresses to full I2P
destinations. This cache is thread-safe using RWMutex and grows unbounded
(applications should monitor memory usage for long-running sessions receiving
from many sources).

Hash Resolution Process:

    1. Convert 32-byte hash to base32 (52 characters)
    2. Append ".b32.i2p" suffix
    3. Check cache for existing entry
    4. If not cached, perform NAMING LOOKUP via SAM bridge
    5. Cache successful result
    6. Return full I2P destination

Example usage:

    resolver := NewHashResolver(sam)
    dest, err := resolver.ResolveHash(hashBytes)
    if err != nil {
        log.Error("Resolution failed:", err)
    }

#### func  NewHashResolver

```go
func NewHashResolver(sam *common.SAM) *HashResolver
```
NewHashResolver creates a new hash resolver with empty cache. The resolver uses
the provided SAM connection for NAMING LOOKUP operations when cache misses
occur.

Example usage:

    resolver := NewHashResolver(sam)

#### func (*HashResolver) CacheSize

```go
func (r *HashResolver) CacheSize() int
```
CacheSize returns the current number of cached entries. This is useful for
monitoring memory usage and cache effectiveness.

Example usage:

    size := resolver.CacheSize()
    log.Info("Cache contains", size, "entries")

#### func (*HashResolver) Clear

```go
func (r *HashResolver) Clear()
```
Clear removes all cached entries. This is useful for testing, memory management
in long-running sessions, or when you want to force fresh NAMING LOOKUP
operations.

Applications with memory constraints may want to implement periodic cache
clearing or LRU eviction policies on top of this basic cache.

Example usage:

    // Clear cache after processing batch
    resolver.Clear()

    // Or clear periodically
    ticker := time.NewTicker(1 * time.Hour)
    go func() {
        for range ticker.C {
            resolver.Clear()
        }
    }()

#### func (*HashResolver) GetCached

```go
func (r *HashResolver) GetCached(hash []byte) (i2pkeys.I2PAddr, bool)
```
GetCached returns cached destination without performing lookup. This allows
checking if a hash has been previously resolved without triggering a potentially
expensive NAMING LOOKUP operation.

Returns:

    - destination: Full I2P destination if cached
    - found: true if entry exists in cache, false otherwise

This method is useful for applications that want to avoid network I/O and only
use already-resolved destinations. It's also useful for testing cache behavior.

Example usage:

    if dest, ok := resolver.GetCached(hash); ok {
        // Use cached destination without network lookup
        writer.SendDatagram(reply, dest)
    } else {
        // Hash not yet resolved - decide whether to resolve now
        log.Info("Hash not in cache, resolution required for reply")
    }

#### func (*HashResolver) ResolveHash

```go
func (r *HashResolver) ResolveHash(hash []byte) (i2pkeys.I2PAddr, error)
```
ResolveHash converts a 32-byte hash to a full I2P destination using NAMING
LOOKUP.

Process:

    1. Validate hash is exactly 32 bytes
    2. Convert to b32.i2p address (base32 encoding + suffix)
    3. Check cache for existing result
    4. If cached, return immediately (fast path)
    5. If not cached, perform NAMING LOOKUP (slow path, network I/O)
    6. Cache successful result for future lookups
    7. Return full destination

This is an expensive operation on cache misses due to network round-trip to I2P
router. Applications should minimize unnecessary resolutions by caching at
application level or reusing the same session resolver.

Error conditions:

    - Invalid hash length (not 32 bytes)
    - Base32 encoding failure (malformed hash)
    - NAMING LOOKUP failure (hash not resolvable, network error, etc.)

Example usage:

    dest, err := resolver.ResolveHash(datagram.SourceHash)
    if err != nil {
        log.Error("Failed to resolve hash:", err)
        return err
    }
    // dest contains full I2P destination (resolved!)
    writer.SendDatagram(reply, dest)

#### type SAM

```go
type SAM struct {
	*common.SAM
}
```

SAM wraps common.SAM to provide datagram3-specific functionality for I2P
messaging. This type extends the base SAM functionality with methods
specifically designed for DATAGRAM3 communication, providing repliable but
datagrams with hash-based source identification.

DATAGRAM3 uses 32-byte hashes instead of full destinations for source
identification, reducing overhead at the cost of full destination verification.
Applications requiring source authentication MUST implement their own
authentication layer.

Example usage: sam := &SAM{SAM: baseSAM}; session, err :=
sam.NewDatagram3Session(id, keys, options)

#### func (*SAM) NewDatagram3Session

```go
func (s *SAM) NewDatagram3Session(id string, keys i2pkeys.I2PKeys, options []string) (*Datagram3Session, error)
```
NewDatagram3Session creates a new repliable but hash-based datagram3 session.
This method establishes a new DATAGRAM3 session for UDP-like messaging over I2P
with hash-based source identification. Session creation can take 2-5 minutes due
to I2P tunnel establishment, so generous timeouts are recommended.

DATAGRAM3 provides repliable datagrams with minimal overhead by using hash-based
source identification instead of full with full destinations destinations.
Received datagrams contain a 32-byte hash that must be resolved via NAMING
LOOKUP to reply. The session maintains a cache to avoid repeated lookups.

Key differences from DATAGRAM and DATAGRAM2:

    - Repliable: Can reply to sender (like DATAGRAM/DATAGRAM2)
    - Unwith full destinations: Source uses hash-based identification (unlike DATAGRAM/DATAGRAM2)
    - Hash-based: Source is 32-byte hash, NOT full destination
    - Lower overhead: Hash-based identification required
    - Reply requires NAMING LOOKUP: Hash must be resolved to full destination

Example usage:

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    session, err := sam.NewDatagram3Session("my-session", keys, []string{"inbound.length=1"})

#### func (*SAM) NewDatagram3SessionWithPorts

```go
func (s *SAM) NewDatagram3SessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*Datagram3Session, error)
```
NewDatagram3SessionWithPorts creates a new datagram3 session with port
specifications. This method allows configuring specific I2CP port ranges for the
session, enabling fine-grained control over network communication ports for
advanced routing scenarios. Port configuration is useful for applications
requiring specific port mappings or PRIMARY session subsessions. This function
automatically creates a UDP listener for SAMv3 UDP forwarding (required for v3
mode).

The FROM_PORT and TO_PORT parameters specify I2CP ports for protocol-level
communication, distinct from the UDP forwarding port which is auto-assigned by
the OS.

Example usage:

    session, err := sam.NewDatagram3SessionWithPorts(id, "8080", "8081", keys, options)

#### func (*SAM) NewDatagram3SessionWithSignature

```go
func (s *SAM) NewDatagram3SessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*Datagram3Session, error)
```
NewDatagram3SessionWithSignature creates a new datagram3 session with custom
signature type. This method allows specifying a custom cryptographic signature
type for the session, enabling advanced security configurations beyond the
default Ed25519 algorithm. DATAGRAM3 supports offline signatures, allowing
pre-signed destinations for enhanced privacy and key management flexibility.

Different signature types provide various security levels for the local
destination:

    - Ed25519 (type 7) - Recommended for most applications
    - ECDSA (types 1-3) - Legacy compatibility
    - RedDSA (type 11) - Advanced privacy features

Example usage:

    session, err := sam.NewDatagram3SessionWithSignature(id, keys, options, "EdDSA_SHA512_Ed25519")
