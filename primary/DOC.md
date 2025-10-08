# primary
--
    import "github.com/go-i2p/go-sam-go/primary"


## Usage

#### type Datagram3SubSession

```go
type Datagram3SubSession struct {
	*datagram3.Datagram3Session
}
```

Datagram3SubSession wraps a datagram3.Datagram3Session to implement the
SubSession interface. This adapter allows Datagram3Session instances to be
managed by primary sessions while maintaining their full functionality and
thread-safe operations.

⚠️ SECURITY WARNING: DATAGRAM3 sources are NOT authenticated and can be spoofed!
⚠️ This sub-session type uses hash-based source identification which is
unauthenticated. ⚠️ Do not trust source addresses without additional
application-level authentication. ⚠️ If you need authenticated sources, use
DatagramSubSession (DATAGRAM) instead.

#### func  NewDatagram3SubSession

```go
func NewDatagram3SubSession(id string, session *datagram3.Datagram3Session) *Datagram3SubSession
```
NewDatagram3SubSession creates a Datagram3SubSession wrapper around a
Datagram3Session. This constructor initializes the wrapper with proper
identification and state management to enable primary session integration.

⚠️ SECURITY WARNING: Sources are UNAUTHENTICATED and can be spoofed!

#### func (*Datagram3SubSession) Active

```go
func (s *Datagram3SubSession) Active() bool
```
Active returns whether this datagram3 sub-session is currently active.

#### func (*Datagram3SubSession) Close

```go
func (s *Datagram3SubSession) Close() error
```
Close closes the datagram3 sub-session and marks it as inactive.

#### func (*Datagram3SubSession) ID

```go
func (s *Datagram3SubSession) ID() string
```
ID returns the unique identifier for this datagram3 sub-session.

#### func (*Datagram3SubSession) Type

```go
func (s *Datagram3SubSession) Type() string
```
Type returns the session type identifier for datagram3 sessions. Returns
"DATAGRAM3" to distinguish from authenticated DATAGRAM sessions.

#### type DatagramSubSession

```go
type DatagramSubSession struct {
	*datagram.DatagramSession
}
```

DatagramSubSession wraps a datagram.DatagramSession to implement the SubSession
interface. This adapter allows DatagramSession instances to be managed by
primary sessions while maintaining their full functionality and thread-safe
operations.

#### func  NewDatagramSubSession

```go
func NewDatagramSubSession(id string, session *datagram.DatagramSession) *DatagramSubSession
```
NewDatagramSubSession creates a DatagramSubSession wrapper around a
DatagramSession. This constructor initializes the wrapper with proper
identification and state management to enable primary session integration.

#### func (*DatagramSubSession) Active

```go
func (s *DatagramSubSession) Active() bool
```
Active returns whether this datagram sub-session is currently active.

#### func (*DatagramSubSession) Close

```go
func (s *DatagramSubSession) Close() error
```
Close closes the datagram sub-session and marks it as inactive.

#### func (*DatagramSubSession) ID

```go
func (s *DatagramSubSession) ID() string
```
ID returns the unique identifier for this datagram sub-session.

#### func (*DatagramSubSession) Type

```go
func (s *DatagramSubSession) Type() string
```
Type returns the session type identifier for datagram sessions.

#### type PrimarySession

```go
type PrimarySession struct {
	*common.BaseSession
}
```

PrimarySession provides master session capabilities for managing multiple
sub-sessions of different types (stream, datagram, raw) within a single I2P
session context. It enables complex applications with multiple communication
patterns while sharing the same I2P identity and tunnel infrastructure for
enhanced efficiency and anonymity.

The primary session manages the lifecycle of all sub-sessions, ensures proper
cleanup cascading when the primary session is closed, and provides thread-safe
operations for creating, managing, and terminating sub-sessions across different
protocols.

#### func  NewPrimarySession

```go
func NewPrimarySession(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error)
```
NewPrimarySession creates a new primary session with the provided SAM
connection, session ID, cryptographic keys, and configuration options. The
primary session acts as a master container that can create and manage multiple
sub-sessions of different types while sharing the same I2P identity and tunnel
infrastructure.

The session uses PRIMARY session type in the SAM protocol, which allows multiple
sub-sessions to be created using the same underlying I2P destination and keys.
This provides better resource efficiency and maintains consistent identity
across different communication patterns within the same application.

Example usage:

    session, err := NewPrimarySession(sam, "my-primary", keys, []string{"inbound.length=2"})
    streamSub, err := session.NewStreamSubSession("stream-1", streamOptions)
    datagramSub, err := session.NewDatagramSubSession("datagram-1", datagramOptions)

#### func  NewPrimarySessionWithSignature

```go
func NewPrimarySessionWithSignature(sam *common.SAM, id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error)
```
NewPrimarySessionWithSignature creates a new primary session with the specified
signature type. This is a package-level function that provides direct access to
signature-aware session creation without requiring wrapper types. It delegates
to the common package for session creation while maintaining the same primary
session functionality and sub-session management capabilities.

The signature type allows specifying custom cryptographic parameters for
enhanced security or compatibility with specific I2P network configurations.
Different signature types provide various security levels, performance
characteristics, and compatibility options.

Example usage:

    session, err := NewPrimarySessionWithSignature(sam, "secure-primary", keys, options, "EdDSA_SHA512_Ed25519")
    streamSub, err := session.NewStreamSubSession("stream-1", streamOptions)

#### func (*PrimarySession) Addr

```go
func (p *PrimarySession) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P address of this primary session. This address represents
the session's identity on the I2P network and is shared by all sub-sessions
created from this primary session. The address is derived from the primary
session's cryptographic keys and remains constant.

Example usage:

    addr := primary.Addr()
    fmt.Printf("Primary session address: %s", addr.Base32())

#### func (*PrimarySession) Close

```go
func (p *PrimarySession) Close() error
```
Close closes the primary session and all associated sub-sessions. This method
performs a complete cleanup cascade, ensuring that all resources are properly
released and all sub-sessions are terminated before closing the primary session
itself. It's safe to call multiple times.

The method first closes all registered sub-sessions, then closes the primary
session's registry and base session. This prevents resource leaks and ensures
proper cleanup of the entire session hierarchy.

Example usage:

    defer primary.Close()

#### func (*PrimarySession) CloseSubSession

```go
func (p *PrimarySession) CloseSubSession(id string) error
```
CloseSubSession closes and unregisters a specific sub-session by its ID. This
method provides selective termination of sub-sessions without affecting the
primary session or other sub-sessions. The sub-session is properly cleaned up
and removed from the registry after closure.

Example usage:

    err := primary.CloseSubSession("stream-1")
    if err != nil {
        log.Printf("Failed to close sub-session: %v", err)
    }

#### func (*PrimarySession) GetSubSession

```go
func (p *PrimarySession) GetSubSession(id string) (SubSession, error)
```
GetSubSession retrieves a sub-session by its unique identifier. Returns the
sub-session instance if found, or an error if the sub-session does not exist or
the primary session is closed. This method provides safe access to registered
sub-sessions for management and operation.

Example usage:

    subSession, err := primary.GetSubSession("stream-1")
    if streamSub, ok := subSession.(*StreamSubSession); ok {
        conn, err := streamSub.Dial("destination.b32.i2p")
    }

#### func (*PrimarySession) ListSubSessions

```go
func (p *PrimarySession) ListSubSessions() []SubSession
```
ListSubSessions returns a list of all currently active sub-sessions. This method
provides a snapshot of all registered sub-sessions that can be safely iterated
without holding locks. The returned list includes sub-sessions of all types
(stream, datagram, raw) currently managed by this primary session.

Example usage:

    subSessions := primary.ListSubSessions()
    for _, sub := range subSessions {
        log.Printf("Sub-session %s (type: %s) is active: %v", sub.ID(), sub.Type(), sub.Active())
    }

#### func (*PrimarySession) NewDatagram3SubSession

```go
func (p *PrimarySession) NewDatagram3SubSession(id string, options []string) (*Datagram3SubSession, error)
```
NewDatagram3SubSession creates a new datagram3 sub-session within this primary
session using SAMv3 UDP forwarding. The sub-session shares the primary session's
I2P identity and tunnel infrastructure while providing full Datagram3Session
functionality for repliable but UNAUTHENTICATED datagram communication. Each
sub-session must have a unique identifier within the primary session scope.

⚠️ SECURITY WARNING: DATAGRAM3 sources are NOT authenticated and can be spoofed!
⚠️ Do not trust source addresses without additional application-level
authentication. ⚠️ If you need authenticated sources, use NewDatagramSubSession
(DATAGRAM) instead.

This implementation uses the SAMv3.3 SESSION ADD protocol to properly register
the subsession with the primary session's SAM connection, ensuring compliance
with the I2P SAM protocol specification for PRIMARY session management.

Per SAMv3.3 specification, DATAGRAM3 subsessions REQUIRE UDP forwarding for
proper operation. Received datagrams contain a 32-byte hash instead of full
authenticated destination. Use the session's hash resolver to convert hashes to
destinations for replies.

Example usage:

    datagram3Sub, err := primary.NewDatagram3SubSession("udp3-handler", []string{"FROM_PORT=8080"})
    reader := datagram3Sub.NewReader()
    writer := datagram3Sub.NewWriter()
    // Receive datagram with UNAUTHENTICATED source hash
    dg, err := reader.ReceiveDatagram()
    // Resolve hash to reply (cached by session)
    err = dg.ResolveSource(datagram3Sub)
    err = writer.SendDatagram([]byte("reply"), dg.Source)

#### func (*PrimarySession) NewDatagramSubSession

```go
func (p *PrimarySession) NewDatagramSubSession(id string, options []string) (*DatagramSubSession, error)
```
NewDatagramSubSession creates a new datagram sub-session within this primary
session. The sub-session shares the primary session's I2P identity and tunnel
infrastructure while providing full DatagramSession functionality for UDP-like
authenticated messaging. Each sub-session must have a unique identifier within
the primary session scope.

This implementation uses the SAMv3.3 SESSION ADD protocol to properly register
the subsession with the primary session's SAM connection, ensuring compliance
with the I2P SAM protocol specification for PRIMARY session management.

Per SAMv3.3 specification, DATAGRAM subsessions REQUIRE a PORT parameter. If
PORT is not included in the options, PORT=0 (any port) will be added
automatically.

Example usage:

    datagramSub, err := primary.NewDatagramSubSession("udp-handler", []string{"PORT=8080", "FROM_PORT=8080"})
    writer := datagramSub.NewWriter()
    reader := datagramSub.NewReader()

#### func (*PrimarySession) NewRawSubSession

```go
func (p *PrimarySession) NewRawSubSession(id string, options []string) (*RawSubSession, error)
```
NewRawSubSession creates a new raw sub-session within this primary session using
SAMv3 UDP forwarding. The sub-session shares the primary session's I2P identity
and tunnel infrastructure while providing full RawSession functionality for
unrepliable datagram communication. Each sub-session must have a unique
identifier within the primary session scope.

This implementation uses the SAMv3.3 SESSION ADD protocol to properly register
the subsession with the primary session's SAM connection, ensuring compliance
with the I2P SAM protocol specification for PRIMARY session management.

Per SAMv3.3 specification, RAW subsessions REQUIRE UDP forwarding for proper
operation. V1/V2 TCP control socket reading is no longer supported.

Example usage:

    rawSub, err := primary.NewRawSubSession("raw-sender", []string{"FROM_PORT=8080"})
    writer := rawSub.NewWriter()
    reader := rawSub.NewReader()

#### func (*PrimarySession) NewStreamSubSession

```go
func (p *PrimarySession) NewStreamSubSession(id string, options []string) (*StreamSubSession, error)
```
NewStreamSubSession creates a new stream sub-session within this primary
session. The sub-session shares the primary session's I2P identity and tunnel
infrastructure while providing full StreamSession functionality for TCP-like
reliable connections. Each sub-session must have a unique identifier within the
primary session scope.

This implementation uses the SAMv3.3 SESSION ADD protocol to properly register
the subsession with the primary session's SAM connection, ensuring compliance
with the I2P SAM protocol specification for PRIMARY session management.

Example usage:

    streamSub, err := primary.NewStreamSubSession("tcp-handler", []string{"FROM_PORT=8080"})
    listener, err := streamSub.Listen()
    conn, err := streamSub.Dial("destination.b32.i2p")

#### func (*PrimarySession) NewUniqueStreamSubSession

```go
func (p *PrimarySession) NewUniqueStreamSubSession(s string) (*StreamSubSession, error)
```
NewUniqueStreamSubSession creates a new unique stream sub-session within this
primary session.

#### func (*PrimarySession) SubSessionCount

```go
func (p *PrimarySession) SubSessionCount() int
```
SubSessionCount returns the number of currently active sub-sessions. This method
provides a quick way to check how many sub-sessions are currently managed by
this primary session across all types.

Example usage:

    count := primary.SubSessionCount()
    log.Printf("Primary session managing %d sub-sessions", count)

#### type PrimarySessionError

```go
type PrimarySessionError struct {
	Op  string // The operation that caused the error
	Err string // The error description
}
```

PrimarySessionError represents errors specific to primary session operations. It
provides structured error information with operation context for debugging and
error handling in primary session management scenarios.

#### func (*PrimarySessionError) Error

```go
func (e *PrimarySessionError) Error() string
```
Error implements the error interface for PrimarySessionError. It provides a
formatted error message that includes both the operation context and the
specific error description for clear error reporting.

#### type RawSubSession

```go
type RawSubSession struct {
	*raw.RawSession
}
```

RawSubSession wraps a raw.RawSession to implement the SubSession interface. This
adapter allows RawSession instances to be managed by primary sessions while
maintaining their full functionality and thread-safe operations.

#### func  NewRawSubSession

```go
func NewRawSubSession(id string, session *raw.RawSession) *RawSubSession
```
NewRawSubSession creates a RawSubSession wrapper around a RawSession. This
constructor initializes the wrapper with proper identification and state
management to enable primary session integration.

#### func (*RawSubSession) Active

```go
func (s *RawSubSession) Active() bool
```
Active returns whether this raw sub-session is currently active.

#### func (*RawSubSession) Close

```go
func (s *RawSubSession) Close() error
```
Close closes the raw sub-session and marks it as inactive.

#### func (*RawSubSession) ID

```go
func (s *RawSubSession) ID() string
```
ID returns the unique identifier for this raw sub-session.

#### func (*RawSubSession) Type

```go
func (s *RawSubSession) Type() string
```
Type returns the session type identifier for raw sessions.

#### type SAM

```go
type SAM struct {
	*common.SAM
}
```

SAM wraps common.SAM to provide primary session functionality for creating and
managing master sessions that can contain multiple sub-sessions of different
types. This type extends the base SAM functionality with methods specifically
designed for primary session management, including session creation with various
configuration options and signature types. Example usage: sam := &SAM{SAM:
baseSAM}; session, err := sam.NewPrimarySession(id, keys, options)

#### func (*SAM) NewPrimarySession

```go
func (s *SAM) NewPrimarySession(id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error)
```
NewPrimarySession creates a new primary session with the SAM bridge using
default settings. This method establishes a new primary session for managing
multiple sub-sessions over I2P with the specified session ID, cryptographic
keys, and configuration options. It uses default signature settings and provides
a simple interface for basic primary session needs.

The primary session acts as a master container that can create and manage
multiple sub-sessions of different types (stream, datagram, raw) while sharing
the same I2P identity and tunnel infrastructure for enhanced efficiency and
consistent anonymity properties.

Example usage:

    session, err := sam.NewPrimarySession("my-primary", keys, []string{"inbound.length=2"})
    streamSub, err := session.NewStreamSubSession("stream-1", streamOptions)

#### func (*SAM) NewPrimarySessionWithPorts

```go
func (s *SAM) NewPrimarySessionWithPorts(id, fromPort, toPort string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error)
```
NewPrimarySessionWithPorts creates a new primary session with port
specifications. This method allows configuring specific port ranges for the
session, enabling fine-grained control over network communication ports for
advanced routing scenarios. Port configuration is useful for applications
requiring specific port mappings, firewall compatibility, or integration with
existing network infrastructure and service discovery mechanisms.

The primary session created with port configuration maintains full multi-session
management capabilities while using the specified port parameters for network
communication optimization and compatibility with existing network
configurations or security requirements.

Example usage:

    session, err := sam.NewPrimarySessionWithPorts(id, "8080", "8081", keys, options)
    rawSub, err := session.NewRawSubSession("raw-1", rawOptions)

#### func (*SAM) NewPrimarySessionWithSignature

```go
func (s *SAM) NewPrimarySessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error)
```
NewPrimarySessionWithSignature creates a new primary session with custom
signature type. This method allows specifying a custom cryptographic signature
type for the session, enabling advanced security configurations beyond the
default signature algorithm. Different signature types provide various security
levels, compatibility options, and performance characteristics for different I2P
network requirements.

The primary session created with custom signature maintains the same
multi-session management capabilities while using the specified cryptographic
parameters for enhanced security or compatibility with specific I2P network
configurations.

Example usage:

    session, err := sam.NewPrimarySessionWithSignature(id, keys, options, "EdDSA_SHA512_Ed25519")
    datagramSub, err := session.NewDatagramSubSession("datagram-1", datagramOptions)

#### type StreamSubSession

```go
type StreamSubSession struct {
	*stream.StreamSession
}
```

StreamSubSession wraps a stream.StreamSession to implement the SubSession
interface. This adapter allows StreamSession instances to be managed by primary
sessions while maintaining their full functionality and thread-safe operations.

#### func  NewStreamSubSession

```go
func NewStreamSubSession(id string, session *stream.StreamSession) *StreamSubSession
```
NewStreamSubSession creates a StreamSubSession wrapper around a StreamSession.
This constructor initializes the wrapper with proper identification and state
management to enable primary session integration.

#### func (*StreamSubSession) Active

```go
func (s *StreamSubSession) Active() bool
```
Active returns whether this stream sub-session is currently active.

#### func (*StreamSubSession) Close

```go
func (s *StreamSubSession) Close() error
```
Close closes the stream sub-session and marks it as inactive.

#### func (*StreamSubSession) ID

```go
func (s *StreamSubSession) ID() string
```
ID returns the unique identifier for this stream sub-session.

#### func (*StreamSubSession) Type

```go
func (s *StreamSubSession) Type() string
```
Type returns the session type identifier for stream sessions.

#### type SubSession

```go
type SubSession interface {
	// ID returns the unique identifier for this sub-session
	ID() string
	// Type returns the session type ("STREAM", "DATAGRAM", "DATAGRAM3", "RAW")
	Type() string
	// Close closes the sub-session and releases its resources
	Close() error
	// Active returns whether the sub-session is currently active
	Active() bool
}
```

SubSession represents a generic interface for sub-sessions that can be managed
by a primary session. All sub-session types (stream, datagram, datagram3, raw)
implement this interface to provide unified lifecycle management and
identification.

#### type SubSessionRegistry

```go
type SubSessionRegistry struct {
}
```

SubSessionRegistry manages a collection of sub-sessions with thread-safe access.
It maintains mappings between session IDs and their corresponding session
instances, enabling efficient lookup, registration, and cleanup operations for
primary sessions.

#### func  NewSubSessionRegistry

```go
func NewSubSessionRegistry() *SubSessionRegistry
```
NewSubSessionRegistry creates a new registry for managing sub-sessions. It
initializes the internal data structures needed for thread-safe sub-session
management and returns a ready-to-use registry instance.

#### func (*SubSessionRegistry) Close

```go
func (r *SubSessionRegistry) Close() error
```
Close closes the registry and all registered sub-sessions. This method ensures
proper cleanup of all resources and marks the registry as closed to prevent
further operations. It's safe to call multiple times.

#### func (*SubSessionRegistry) Count

```go
func (r *SubSessionRegistry) Count() int
```
Count returns the number of currently registered sub-sessions. This method is
thread-safe and provides a quick way to check registry size.

#### func (*SubSessionRegistry) Get

```go
func (r *SubSessionRegistry) Get(id string) (SubSession, bool)
```
Get retrieves a sub-session by ID from the registry. Returns the session
instance and true if found, or nil and false if not found. This method is
thread-safe and provides read-only access to registered sessions.

#### func (*SubSessionRegistry) IsClosed

```go
func (r *SubSessionRegistry) IsClosed() bool
```
IsClosed returns whether the registry has been closed. This method is
thread-safe and can be used to check registry state.

#### func (*SubSessionRegistry) List

```go
func (r *SubSessionRegistry) List() []SubSession
```
List returns a copy of all currently registered sub-sessions. This method is
thread-safe and returns a snapshot of the registry state that can be safely
iterated without holding locks.

#### func (*SubSessionRegistry) Register

```go
func (r *SubSessionRegistry) Register(id string, session SubSession) error
```
Register adds a sub-session to the registry with the specified ID. Returns an
error if the registry is closed or if a session with the same ID already exists.
This method is thread-safe and can be called concurrently.

#### func (*SubSessionRegistry) Unregister

```go
func (r *SubSessionRegistry) Unregister(id string) error
```
Unregister removes a sub-session from the registry by ID. Returns an error if
the registry is closed or if no session with the specified ID exists. This
method is thread-safe and can be called concurrently.
