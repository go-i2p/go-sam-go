# sam3
--
    import "github.com/go-i2p/go-sam-go"

Package sam3 provides configuration wrapper functions for exposing common
package configuration functionality at the root level. This file implements
wrapper functions for SAMEmit options and I2PConfig creation that enable
applications to configure I2P tunnel parameters and session behaviors through a
clean, sam3-compatible API.

Package sam provides a pure-Go implementation of SAMv3.3 (Simple Anonymous
Messaging) for I2P networks. This is the root package wrapper that provides
sam3-compatible API surface while delegating implementation details to
specialized sub-packages.

Package sam provides a pure-Go implementation of SAMv3.3 for I2P networks. This
file implements the main wrapper functions that delegate to sub-package
implementations while providing the sam3-compatible API surface at the root
package level.

Package sam provides type aliases and wrappers for exposing sub-package types at
the root level. This file implements the sam3-compatible API surface by creating
type aliases that delegate to the appropriate sub-package implementations while
maintaining a clean public interface.

## Usage

```go
const (
	// Sig_NONE is deprecated, use Sig_EdDSA_SHA512_Ed25519 instead for secure signatures.
	Sig_NONE = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"

	// Sig_DSA_SHA1 specifies DSA with SHA1 signature type (legacy, not recommended for new applications).
	Sig_DSA_SHA1 = "SIGNATURE_TYPE=DSA_SHA1"

	// Sig_ECDSA_SHA256_P256 specifies ECDSA with SHA256 on P256 curve signature type.
	Sig_ECDSA_SHA256_P256 = "SIGNATURE_TYPE=ECDSA_SHA256_P256"

	// Sig_ECDSA_SHA384_P384 specifies ECDSA with SHA384 on P384 curve signature type.
	Sig_ECDSA_SHA384_P384 = "SIGNATURE_TYPE=ECDSA_SHA384_P384"

	// Sig_ECDSA_SHA512_P521 specifies ECDSA with SHA512 on P521 curve signature type.
	Sig_ECDSA_SHA512_P521 = "SIGNATURE_TYPE=ECDSA_SHA512_P521"

	// Sig_EdDSA_SHA512_Ed25519 specifies EdDSA with SHA512 on Ed25519 curve signature type (recommended).
	Sig_EdDSA_SHA512_Ed25519 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"

	// Sig_DEFAULT points to the recommended secure signature type for new applications.
	Sig_DEFAULT = Sig_EdDSA_SHA512_Ed25519
)
```
Signature type constants for I2P destination key generation. These specify the
cryptographic signature algorithm used for I2P destinations. SIG_DEFAULT points
to the recommended secure signature type for new applications.

```go
var (
	// Options_Humongous provides maximum anonymity and redundancy for extremely high-value traffic.
	// Suitable for applications requiring the highest level of anonymity protection with significant
	// resource overhead. Uses 3-hop tunnels with high redundancy and parallel connections.
	Options_Humongous = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=3", "outbound.backupQuantity=3",
		"inbound.quantity=6", "outbound.quantity=6",
	}

	// Options_Large provides strong anonymity for high-traffic applications.
	// Suitable for applications shuffling large amounts of traffic with good anonymity protection.
	// Balances performance and anonymity with reasonable resource usage.
	Options_Large = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=1", "outbound.backupQuantity=1",
		"inbound.quantity=4", "outbound.quantity=4",
	}

	// Options_Wide provides minimal anonymity but high performance for traffic requiring low latency.
	// Suitable for applications prioritizing speed over anonymity. Uses 1-hop tunnels with
	// moderate redundancy for basic privacy protection while maintaining performance.
	Options_Wide = []string{
		"inbound.length=1", "outbound.length=1",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=2", "outbound.backupQuantity=2",
		"inbound.quantity=3", "outbound.quantity=3",
	}

	// Options_Medium provides balanced anonymity and performance for moderate traffic loads.
	// Suitable for applications with medium traffic requirements that need good anonymity
	// without excessive resource overhead. Provides solid baseline protection.
	Options_Medium = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=2", "outbound.quantity=2",
	}

	// Options_Default provides sensible defaults for most use cases.
	// Recommended starting point for most applications. Provides good anonymity
	// with reasonable performance characteristics and resource usage.
	Options_Default = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=0", "outbound.lengthVariance=0",
		"inbound.backupQuantity=1", "outbound.backupQuantity=1",
		"inbound.quantity=1", "outbound.quantity=1",
	}

	// Options_Small provides basic anonymity for low-traffic, short-duration connections.
	// Suitable only for applications with minimal traffic requirements and short connection
	// lifetimes. Offers basic privacy with minimal resource usage.
	Options_Small = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=1", "outbound.quantity=1",
	}

	// Options_Warning_ZeroHop disables all anonymization - zero hop configuration.
	// WARNING: This configuration provides NO anonymity protection and should only be used
	// for testing or debugging purposes. All traffic is directly routed without tunnel protection.
	Options_Warning_ZeroHop = []string{
		"inbound.length=0", "outbound.length=0",
		"inbound.lengthVariance=0", "outbound.lengthVariance=0",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=2", "outbound.quantity=2",
	}
)
```
Predefined tunnel configuration option sets for different traffic patterns and
anonymity requirements. These option sets balance performance, anonymity, and
resource usage for common use cases. Each option set specifies tunnel length,
variance, backup quantity, and parallel tunnel count.

```go
var (
	// SAM_HOST specifies the SAM bridge host address.
	// Can be overridden with the 'sam_host' environment variable.
	// Defaults to localhost (127.0.0.1) for local I2P router connections.
	SAM_HOST = getEnv("sam_host", "127.0.0.1")

	// SAM_PORT specifies the SAM bridge port number.
	// Can be overridden with the 'sam_port' environment variable.
	// Defaults to 7656, the standard SAM bridge port.
	SAM_PORT = getEnv("sam_port", "7656")

	// PrimarySessionSwitch enables primary session functionality.
	// This is used internally to enable multi-session capabilities.
	PrimarySessionSwitch = PrimarySessionString()
)
```
SAM bridge connection variables with environment variable support. These
variables can be overridden using sam_host and sam_port environment variables.

#### func  ExtractDest

```go
func ExtractDest(input string) string
```
ExtractDest extracts the destination address from a SAM protocol response
string. This utility function takes the first space-separated token from the
input as the destination. It's commonly used for parsing SAM session creation
responses and connection messages.

Example:

    dest := ExtractDest("ABCD1234...destination_address RESULT=OK")
    // Returns: "ABCD1234...destination_address"

#### func  ExtractPairInt

```go
func ExtractPairInt(input, value string) int
```
ExtractPairInt extracts an integer value from a key=value pair in a
space-separated string. This utility function searches for the specified key and
converts its value to an integer. Returns 0 if the key is not found or the value
cannot be converted to an integer.

Example:

    port := ExtractPairInt("HOST=example.org PORT=1234 TYPE=stream", "PORT")
    // Returns: 1234

#### func  ExtractPairString

```go
func ExtractPairString(input, value string) string
```
ExtractPairString extracts a string value from a key=value pair in a
space-separated string. This utility function searches for the specified key and
returns its associated value. Returns empty string if the key is not found or
has no value.

Example:

    host := ExtractPairString("HOST=example.org PORT=1234 TYPE=stream", "HOST")
    // Returns: "example.org"

#### func  GenerateOptionString

```go
func GenerateOptionString(opts []string) string
```
GenerateOptionString converts a slice of tunnel options into a single
space-separated string. This utility function takes an array of I2P tunnel
configuration options and formats them for use in SAM protocol commands. Each
option should be in "key=value" format.

Example:

    opts := []string{"inbound.length=3", "outbound.length=3"}
    result := GenerateOptionString(opts)
    // Returns: "inbound.length=3 outbound.length=3"

#### func  GetSAM3Logger

```go
func GetSAM3Logger() *logrus.Logger
```
GetSAM3Logger returns the initialized logger instance used by the SAM library.
This function provides access to the structured logger for applications that
want to integrate with the library's logging system or adjust log levels.

The logger is configured with appropriate fields for I2P and SAM operations,
supporting debug, info, warn, and error levels with structured output.

#### func  IgnorePortError

```go
func IgnorePortError(err error) error
```
IgnorePortError filters out "missing port in address" errors for convenience
when parsing addresses. This utility function is used when working with
addresses that may not include port numbers. Returns nil if the error is about a
missing port, otherwise returns the original error unchanged.

This is particularly useful when parsing I2P destination addresses that don't
always include port specifications, allowing graceful handling of address
parsing operations.

Example:

    _, _, err := net.SplitHostPort("example.i2p")  // This would error
    err = IgnorePortError(err)  // This returns nil

#### func  InitializeSAM3Logger

```go
func InitializeSAM3Logger()
```
InitializeSAM3Logger configures the logging system for the SAM library. This
function sets up the logger with appropriate configuration for I2P operations,
including proper log levels and formatting for SAM protocol debugging.

The logger respects environment variables for configuration: - DEBUG_I2P:
Controls log level (debug, info, warn, error) Applications should call this once
during initialization if they want to enable structured logging for SAM
operations.

#### func  PrimarySessionString

```go
func PrimarySessionString() string
```
PrimarySessionString returns the primary session configuration identifier. This
function provides compatibility with the sam3 library's primary session switch
mechanism for enabling advanced session management features.

#### func  RandString

```go
func RandString() string
```
RandString generates a random string suitable for use as session identifiers or
tunnel names. This utility function creates cryptographically secure random
strings using I2P's random number generator. The generated strings are URL-safe
and suitable for use in SAM protocol commands and session identification.

Returns a random string that can be used for session IDs, tunnel names, or other
identifiers that require uniqueness and unpredictability in I2P operations.

#### func  SAMDefaultAddr

```go
func SAMDefaultAddr(fallforward string) string
```
SAMDefaultAddr constructs the default SAM bridge address with fallback support.
This utility function provides a standardized way to determine the SAM bridge
address, using the provided fallback if the standard environment variables are
not set.

The function checks SAM_HOST and SAM_PORT variables first, then falls back to
the provided fallforward parameter if those are not available. This enables
flexible configuration while providing sensible defaults for most I2P
installations.

Example:

    addr := SAMDefaultAddr("127.0.0.1:7656")
    // Returns: "127.0.0.1:7656" (or values from SAM_HOST/SAM_PORT if set)

#### func  SetAccessList

```go
func SetAccessList(s []string) func(*SAMEmit) error
```
SetAccessList sets the list of destinations for access control. The behavior
depends on the access list type: whitelist allows only these destinations,
blacklist denies these destinations.

Example usage:

    destinations := []string{"dest1.b32.i2p", "dest2.b32.i2p"}
    emit, err := NewEmit(SetAccessList(destinations))

#### func  SetAccessListType

```go
func SetAccessListType(s string) func(*SAMEmit) error
```
SetAccessListType sets the type of access control to apply. Valid values are
"whitelist" (allow only listed destinations), "blacklist" (deny listed
destinations), or "none" (no filtering).

Example usage:

    emit, err := NewEmit(SetAccessListType("whitelist"))

#### func  SetAllowZeroIn

```go
func SetAllowZeroIn(b bool) func(*SAMEmit) error
```
SetAllowZeroIn enables or disables acceptance of zero-hop inbound tunnels.
Zero-hop tunnels provide no anonymity but offer better performance for
applications that don't require anonymity protection.

Example usage:

    emit, err := NewEmit(SetAllowZeroIn(false))

#### func  SetAllowZeroOut

```go
func SetAllowZeroOut(b bool) func(*SAMEmit) error
```
SetAllowZeroOut enables or disables acceptance of zero-hop outbound tunnels.
Zero-hop tunnels provide no anonymity but offer better performance for
applications that don't require anonymity protection.

Example usage:

    emit, err := NewEmit(SetAllowZeroOut(false))

#### func  SetCloseIdle

```go
func SetCloseIdle(b bool) func(*SAMEmit) error
```
SetCloseIdle enables or disables complete tunnel closure during extended idle
periods. When enabled, closes all tunnels after the specified idle time to
completely conserve router resources.

Example usage:

    emit, err := NewEmit(SetCloseIdle(true))

#### func  SetCloseIdleTime

```go
func SetCloseIdleTime(u int) func(*SAMEmit) error
```
SetCloseIdleTime sets the time in minutes to wait before closing all tunnels.
After this period of inactivity, all tunnels will be closed to conserve router
resources.

Example usage:

    emit, err := NewEmit(SetCloseIdleTime(30)) // 30 minutes

#### func  SetCloseIdleTimeMs

```go
func SetCloseIdleTimeMs(u int) func(*SAMEmit) error
```
SetCloseIdleTimeMs sets the time in milliseconds to wait before closing all
tunnels. After this period of inactivity, all tunnels will be closed to conserve
router resources.

Example usage:

    emit, err := NewEmit(SetCloseIdleTimeMs(1800000)) // 30 minutes

#### func  SetCompress

```go
func SetCompress(b bool) func(*SAMEmit) error
```
SetCompress enables or disables data compression for tunnel traffic. Compression
can reduce bandwidth usage but may impact performance and could potentially
affect anonymity through traffic analysis.

Example usage:

    emit, err := NewEmit(SetCompress(true))

#### func  SetEncrypt

```go
func SetEncrypt(b bool) func(*SAMEmit) error
```
SetEncrypt enables or disables encrypted lease sets for enhanced security.
Encrypted lease sets provide additional protection against traffic analysis but
may slightly impact performance and compatibility.

Example usage:

    emit, err := NewEmit(SetEncrypt(true))

#### func  SetFastRecieve

```go
func SetFastRecieve(b bool) func(*SAMEmit) error
```
SetFastRecieve enables or disables fast receive mode for improved performance.
When enabled, bypasses some protocol overhead for faster data transmission at
the potential cost of some reliability guarantees.

Example usage:

    emit, err := NewEmit(SetFastRecieve(true))

#### func  SetInBackups

```go
func SetInBackups(u int) func(*SAMEmit) error
```
SetInBackups sets the number of backup inbound tunnels (0-5). Backup tunnels are
pre-built spare tunnels that can quickly replace failed primary tunnels,
improving connection reliability.

Example usage:

    emit, err := NewEmit(SetInBackups(1))

#### func  SetInLength

```go
func SetInLength(u int) func(*SAMEmit) error
```
SetInLength sets the number of hops for inbound tunnels (0-6). Higher values
provide better anonymity but increase latency and resource usage. Most
applications use 3 hops as a balance between security and performance.

Example usage:

    emit, err := NewEmit(SetInLength(3))

#### func  SetInQuantity

```go
func SetInQuantity(u int) func(*SAMEmit) error
```
SetInQuantity sets the number of inbound tunnels to maintain (1-16). More
tunnels provide better load distribution and redundancy but consume more router
resources. Most applications use 1-4 tunnels.

Example usage:

    emit, err := NewEmit(SetInQuantity(2))

#### func  SetInVariance

```go
func SetInVariance(i int) func(*SAMEmit) error
```
SetInVariance sets the variance for inbound tunnel hop counts (-6 to 6). This
adds randomness to tunnel lengths to prevent traffic analysis. Positive values
increase maximum hops, negative values allow shorter tunnels.

Example usage:

    emit, err := NewEmit(SetInVariance(1))

#### func  SetLeaseSetKey

```go
func SetLeaseSetKey(s string) func(*SAMEmit) error
```
SetLeaseSetKey sets the public key for lease set encryption. This key is used to
encrypt the lease set information, providing additional security for the
destination's routing information.

Example usage:

    emit, err := NewEmit(SetLeaseSetKey("base64-encoded-key"))

#### func  SetLeaseSetPrivateKey

```go
func SetLeaseSetPrivateKey(s string) func(*SAMEmit) error
```
SetLeaseSetPrivateKey sets the private key for lease set decryption. This key is
used to decrypt lease set information when encrypted lease sets are enabled for
enhanced security.

Example usage:

    emit, err := NewEmit(SetLeaseSetPrivateKey("base64-encoded-private-key"))

#### func  SetLeaseSetPrivateSigningKey

```go
func SetLeaseSetPrivateSigningKey(s string) func(*SAMEmit) error
```
SetLeaseSetPrivateSigningKey sets the private signing key for lease set
authentication. This key is used to sign lease set information to ensure
authenticity and prevent tampering with routing information.

Example usage:

    emit, err := NewEmit(SetLeaseSetPrivateSigningKey("base64-encoded-signing-key"))

#### func  SetMessageReliability

```go
func SetMessageReliability(s string) func(*SAMEmit) error
```
SetMessageReliability sets the reliability level for message delivery. Options
include "none", "BestEffort", or "Guaranteed" depending on the application's
reliability requirements and performance trade-offs.

Example usage:

    emit, err := NewEmit(SetMessageReliability("BestEffort"))

#### func  SetName

```go
func SetName(s string) func(*SAMEmit) error
```
SetName sets the tunnel name for identification and debugging purposes. This
name appears in I2P router logs and management interfaces to help identify and
manage specific tunnels created by your application.

Example usage:

    emit, err := NewEmit(SetName("my-app-tunnel"))

#### func  SetOutBackups

```go
func SetOutBackups(u int) func(*SAMEmit) error
```
SetOutBackups sets the number of backup outbound tunnels (0-5). Backup tunnels
are pre-built spare tunnels that can quickly replace failed primary tunnels,
improving connection reliability.

Example usage:

    emit, err := NewEmit(SetOutBackups(1))

#### func  SetOutLength

```go
func SetOutLength(u int) func(*SAMEmit) error
```
SetOutLength sets the number of hops for outbound tunnels (0-6). Higher values
provide better anonymity but increase latency and resource usage. Most
applications use 3 hops as a balance between security and performance.

Example usage:

    emit, err := NewEmit(SetOutLength(3))

#### func  SetOutQuantity

```go
func SetOutQuantity(u int) func(*SAMEmit) error
```
SetOutQuantity sets the number of outbound tunnels to maintain (1-16). More
tunnels provide better load distribution and redundancy but consume more router
resources. Most applications use 1-4 tunnels.

Example usage:

    emit, err := NewEmit(SetOutQuantity(2))

#### func  SetOutVariance

```go
func SetOutVariance(i int) func(*SAMEmit) error
```
SetOutVariance sets the variance for outbound tunnel hop counts (-6 to 6). This
adds randomness to tunnel lengths to prevent traffic analysis. Positive values
increase maximum hops, negative values allow shorter tunnels.

Example usage:

    emit, err := NewEmit(SetOutVariance(1))

#### func  SetReduceIdle

```go
func SetReduceIdle(b bool) func(*SAMEmit) error
```
SetReduceIdle enables or disables tunnel reduction during extended idle periods.
When enabled, reduces the number of active tunnels during idle time to conserve
router resources while maintaining minimal connectivity.

Example usage:

    emit, err := NewEmit(SetReduceIdle(true))

#### func  SetReduceIdleQuantity

```go
func SetReduceIdleQuantity(u int) func(*SAMEmit) error
```
SetReduceIdleQuantity sets the minimum number of tunnels during idle periods.
When idle reduction is enabled, this is the number of tunnels that will be
maintained during periods of low activity.

Example usage:

    emit, err := NewEmit(SetReduceIdleQuantity(1))

#### func  SetReduceIdleTime

```go
func SetReduceIdleTime(u int) func(*SAMEmit) error
```
SetReduceIdleTime sets the time in minutes to wait before reducing tunnels.
After this period of inactivity, the number of tunnels will be reduced to the
quantity specified by SetReduceIdleQuantity.

Example usage:

    emit, err := NewEmit(SetReduceIdleTime(10)) // 10 minutes

#### func  SetReduceIdleTimeMs

```go
func SetReduceIdleTimeMs(u int) func(*SAMEmit) error
```
SetReduceIdleTimeMs sets the time in milliseconds to wait before reducing
tunnels. After this period of inactivity, the number of tunnels will be reduced
to the quantity specified by SetReduceIdleQuantity.

Example usage:

    emit, err := NewEmit(SetReduceIdleTimeMs(600000)) // 10 minutes

#### func  SetSAMAddress

```go
func SetSAMAddress(s string) func(*SAMEmit) error
```
SetSAMAddress sets the SAM bridge address all-at-once using "host:port" format.
This convenience function parses the address string and sets both host and port
simultaneously for simplified SAM bridge configuration.

Example usage:

    emit, err := NewEmit(SetSAMAddress("127.0.0.1:7656"))

#### func  SetSAMHost

```go
func SetSAMHost(s string) func(*SAMEmit) error
```
SetSAMHost sets the hostname or IP address of the SAM bridge. The SAM bridge is
the interface that allows applications to communicate with the I2P router for
creating sessions and managing connections.

Example usage:

    emit, err := NewEmit(SetSAMHost("127.0.0.1"))

#### func  SetSAMPort

```go
func SetSAMPort(s string) func(*SAMEmit) error
```
SetSAMPort sets the port number of the SAM bridge using a string value. The port
must be a valid TCP port number (0-65535) where the I2P router's SAM bridge is
listening for incoming connections.

Example usage:

    emit, err := NewEmit(SetSAMPort("7656"))

#### func  SetType

```go
func SetType(s string) func(*SAMEmit) error
```
SetType sets the session type for the forwarder server. Valid values are
"STREAM", "DATAGRAM", or "RAW" corresponding to different communication patterns
available in the I2P network protocol.

Example usage:

    emit, err := NewEmit(SetType("STREAM"))

#### func  SplitHostPort

```go
func SplitHostPort(hostport string) (string, string, error)
```
SplitHostPort separates host and port from a combined address string with
I2P-aware handling. Unlike net.SplitHostPort, this function handles I2P
addresses gracefully, including those without explicit port specifications.
Returns host, port as strings, and error.

This function is I2P-aware and handles the common case where I2P destination
addresses don't include port numbers. Port defaults to "0" if not specified, and
the function uses IgnorePortError internally to handle missing port situations
gracefully.

Example:

    host, port, err := SplitHostPort("example.i2p")
    // Returns: "example.i2p", "0", nil

#### type BaseSession

```go
type BaseSession = common.BaseSession
```

BaseSession represents the underlying session functionality that all session
types extend. It provides common operations like connection management, key
access, and standard net.Conn interface implementation for I2P session
operations.

#### type DatagramSession

```go
type DatagramSession = datagram.DatagramSession
```

DatagramSession provides UDP-like messaging capabilities with message
reliability and ordering guarantees over I2P networks. It handles signed,
authenticated messaging with replay protection for secure datagram communication
patterns.

#### type I2PConfig

```go
type I2PConfig = common.I2PConfig
```

I2PConfig manages I2P tunnel configuration options including tunnel length,
backup quantities, variance settings, and other I2P-specific parameters.
Applications use this to customize their anonymity and performance
characteristics.

#### func  NewConfig

```go
func NewConfig(opts ...func(*I2PConfig) error) (*I2PConfig, error)
```
NewConfig creates a new I2PConfig instance with default values and applies
functional options. Returns a configured instance ready for use in session
creation or an error if any option fails. This delegates to common.NewConfig
while providing the sam3 API surface.

Example usage:

    config, err := NewConfig(SetInLength(4), SetOutLength(4))
    if err != nil {
        log.Fatal("Failed to create config:", err)
    }

#### type Option

```go
type Option = common.Option
```

Option represents a functional option for configuring SAMEmit instances. This
follows the functional options pattern to provide type-safe configuration with
clear error handling and composable session parameter management.

#### type Options

```go
type Options = common.Options
```

Options represents a map of configuration options that can be applied to
sessions. This provides a flexible way to specify tunnel parameters and session
behaviors using key-value pairs that are converted to SAM protocol commands.

#### type PrimarySession

```go
type PrimarySession = primary.PrimarySession
```

PrimarySession provides master session capabilities that can create and manage
multiple sub-sessions of different types (stream, datagram, raw) within a single
I2P session context. This enables complex applications with multiple
communication patterns while sharing the same I2P identity and tunnel
infrastructure.

#### type RawSession

```go
type RawSession = raw.RawSession
```

RawSession provides unrepliable datagram communication over I2P networks.
Messages are encrypted end-to-end but senders cannot be identified or replied
to, providing the highest level of sender anonymity for one-way communication.

#### type SAM

```go
type SAM struct {
	*common.SAM
}
```

SAM represents the core SAM bridge connection and provides methods for session
creation and I2P address resolution. It embeds common.SAM to enable method
extension while exposing the sam3-compatible interface at the root package
level.

#### func  NewSAM

```go
func NewSAM(address string) (*SAM, error)
```
NewSAM creates a new SAM connection to the specified address and performs the
initial handshake. This is the main entry point for establishing connections to
the I2P SAM bridge. Address should be in the format "host:port", typically
"127.0.0.1:7656" for local I2P routers.

The function connects to the SAM bridge, performs the protocol handshake, and
initializes the resolver for I2P name lookups. It returns a ready-to-use SAM
instance or an error if any step of the initialization process fails.

Example:

    sam, err := NewSAM("127.0.0.1:7656")
    if err != nil {
        log.Fatal(err)
    }
    defer sam.Close()

#### func (*SAM) NewDatagramSession

```go
func (sam *SAM) NewDatagramSession(id string, keys i2pkeys.I2PKeys, options []string, udpPort int) (*DatagramSession, error)
```
NewDatagramSession creates a new datagram session for UDP-like authenticated
messaging over I2P. Datagram sessions provide connectionless communication with
message authentication and replay protection, suitable for applications that
need fast, lightweight messaging without the overhead of connection
establishment.

Unlike raw sessions, datagram sessions provide sender authentication and message
integrity verification, making them suitable for applications where message
authenticity is important. Each message includes cryptographic signatures that
allow recipients to verify the sender's identity and detect tampering.

Returns a DatagramSession ready for sending and receiving authenticated
datagrams, or an error if the session creation fails.

The udpPort parameter specifies the local UDP port for the session's datagram
interface. If set to 0, the I2P router will use the standard SAM UDP port
(typically 7655). This port is used for the local UDP socket that communicates
with the I2P datagram subsystem for message forwarding and reception.

Example usage:

    sam, err := NewSAM("127.0.0.1:7656")
    keys, _ := i2pkeys.NewKeys(i2pkeys.KT_ECDSA_SHA256_P256)
    session, err := sam.NewDatagramSession("chat-app", keys, Options_Medium, 0)

#### func (*SAM) NewPrimarySession

```go
func (sam *SAM) NewPrimarySession(id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error)
```
NewPrimarySession creates a new primary session that can manage multiple
sub-sessions of different types (stream, datagram, raw) within a single I2P
session context. The primary session enables complex applications with multiple
communication patterns while sharing the same I2P identity and tunnel
infrastructure for enhanced efficiency.

The session ID must be unique and will be used to identify this session in the
SAM protocol. The I2P keys define the cryptographic identity that will be shared
across all sub-sessions created from this primary session. Configuration options
control tunnel parameters such as length, backup quantity, and other
I2P-specific settings.

Returns a PrimarySession that can create and manage sub-sessions, or an error if
the session creation fails due to SAM protocol errors, network issues, or
invalid configuration parameters.

Example usage:

    sam, err := NewSAM("127.0.0.1:7656")
    keys, _ := i2pkeys.NewKeys(i2pkeys.KT_ECDSA_SHA256_P256)
    primary, err := sam.NewPrimarySession("app-primary", keys, Options_Default)
    streamSub, err := primary.NewStreamSubSession("tcp-handler", []string{})
    datagramSub, err := primary.NewDatagramSubSession("udp-handler", []string{})

#### func (*SAM) NewPrimarySessionWithSignature

```go
func (sam *SAM) NewPrimarySessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error)
```
NewPrimarySessionWithSignature creates a new primary session with a specific
signature type. This method provides the same functionality as NewPrimarySession
but allows explicit control over the cryptographic signature algorithm used for
the session's I2P identity.

The signature type must be one of the supported I2P signature algorithms (use
the Sig_* constants defined in this package). Different signature types offer
different security levels and performance characteristics - EdDSA_SHA512_Ed25519
is recommended for most applications as it provides strong security with good
performance.

Returns a PrimarySession configured with the specified signature type, or an
error if the session creation fails or the signature type is not supported by
the I2P router.

Example usage:

    sam, err := NewSAM("127.0.0.1:7656")
    keys, _ := i2pkeys.NewKeys(i2pkeys.KT_EdDSA_SHA512_Ed25519)
    primary, err := sam.NewPrimarySessionWithSignature("secure-primary", keys,
    	Options_Default, Sig_EdDSA_SHA512_Ed25519)

#### func (*SAM) NewRawSession

```go
func (sam *SAM) NewRawSession(id string, keys i2pkeys.I2PKeys, options []string, udpPort int) (*RawSession, error)
```
NewRawSession creates a new raw session for unrepliable datagram communication
over I2P. Raw sessions provide the most lightweight form of I2P communication,
where messages are encrypted end-to-end but do not include sender authentication
or guarantee delivery. Recipients cannot identify the sender or send replies
directly.

Raw sessions are suitable for applications that need maximum anonymity and
minimal overhead, such as anonymous publishing, voting systems, or situations
where sender identity must remain completely hidden. The lack of sender
authentication provides stronger anonymity but eliminates the ability to verify
message authenticity.

The udpPort parameter specifies the local UDP port for the session's datagram
interface. This port is used for the local UDP socket that communicates with the
I2P raw datagram subsystem.

Returns a RawSession ready for sending anonymous unrepliable datagrams, or an
error if the session creation fails.

Example usage:

    sam, err := NewSAM("127.0.0.1:7656")
    keys, _ := i2pkeys.NewKeys(i2pkeys.KT_ECDSA_SHA256_P256)
    session, err := sam.NewRawSession("anonymous-publisher", keys, Options_Small, 0)
    writer := session.NewWriter()
    reader := session.NewReader()

#### func (*SAM) NewStreamSession

```go
func (sam *SAM) NewStreamSession(id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error)
```
NewStreamSession creates a new stream session for TCP-like reliable connections
over I2P. Stream sessions provide connection-oriented communication with
guarantees for message ordering, delivery, and flow control, similar to TCP but
routed through I2P tunnels.

The session can be used to create listeners for accepting incoming connections
or to establish outbound connections to other I2P destinations. All connections
share the same I2P identity defined by the provided keys and benefit from the
tunnel configuration specified in the options.

Returns a StreamSession ready for creating connections, or an error if the
session creation fails due to SAM protocol errors, network issues, or invalid
configuration.

Example usage:

    sam, err := NewSAM("127.0.0.1:7656")
    keys, _ := i2pkeys.NewKeys(i2pkeys.KT_ECDSA_SHA256_P256)
    session, err := sam.NewStreamSession("web-server", keys, Options_Default)
    listener, err := session.Listen()
    conn, err := session.Dial("destination.b32.i2p")

#### func (*SAM) NewStreamSessionWithSignature

```go
func (sam *SAM) NewStreamSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error)
```
NewStreamSessionWithSignature creates a new stream session with a specific
signature type. This method provides the same functionality as NewStreamSession
but allows explicit control over the cryptographic signature algorithm used for
the session's I2P identity.

The signature type determines the cryptographic strength and performance
characteristics of the session. EdDSA_SHA512_Ed25519 is recommended for most
applications, while ECDSA variants may be preferred for compatibility with older
I2P routers or specific security requirements.

Returns a StreamSession configured with the specified signature type, or an
error if the session creation fails or the signature type is not supported.

Example usage:

    sam, err := NewSAM("127.0.0.1:7656")
    keys, _ := i2pkeys.NewKeys(i2pkeys.KT_EdDSA_SHA512_Ed25519)
    session, err := sam.NewStreamSessionWithSignature("secure-stream", keys,
    	Options_Large, Sig_EdDSA_SHA512_Ed25519)

#### func (*SAM) NewStreamSessionWithSignatureAndPorts

```go
func (sam *SAM) NewStreamSessionWithSignatureAndPorts(id, from, to string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error)
```
NewStreamSessionWithSignatureAndPorts creates a new stream session with
signature type and specific port configuration. This method enables advanced
port mapping scenarios where the session should bind to specific local ports or
forward to specific remote ports.

The 'from' parameter specifies the local port or port range that the session
should use for incoming connections. The 'to' parameter specifies the target
port or port range for outbound connections. Port specifications can be single
ports ("80") or ranges ("8080-8090").

This method is particularly useful for applications that need to maintain
consistent port mappings or integrate with existing network infrastructure that
expects specific port configurations.

Returns a StreamSession with the specified port configuration, or an error if
the session creation fails or the port configuration is invalid.

Example usage:

    sam, err := NewSAM("127.0.0.1:7656")
    keys, _ := i2pkeys.NewKeys(i2pkeys.KT_ECDSA_SHA256_P256)
    session, err := sam.NewStreamSessionWithSignatureAndPorts("http-proxy",
    	"8080", "80", keys, Options_Default, Sig_ECDSA_SHA256_P256)

#### type SAMConn

```go
type SAMConn = stream.StreamConn
```

SAMConn implements net.Conn for I2P streaming connections, providing a standard
Go networking interface for TCP-like reliable communication over I2P networks.
This alias maintains sam3 API compatibility while delegating to
stream.StreamConn.

#### type SAMEmit

```go
type SAMEmit = common.SAMEmit
```

SAMEmit handles SAM protocol message generation and configuration management. It
embeds I2PConfig to provide comprehensive session configuration capabilities
while managing the underlying SAM protocol communication requirements.

#### func  NewEmit

```go
func NewEmit(opts ...func(*SAMEmit) error) (*SAMEmit, error)
```
NewEmit creates a new SAMEmit instance with the specified configuration options.
Applies functional options to configure the emitter with custom settings.
Returns an error if any option fails to apply correctly.

Example usage:

    emit, err := NewEmit(SetSAMHost("localhost"), SetSAMPort("7656"))
    if err != nil {
        log.Fatal("Failed to create emitter:", err)
    }

#### type SAMResolver

```go
type SAMResolver = common.SAMResolver
```

SAMResolver provides I2P address resolution services through the SAM protocol.
It wraps the common.SAMResolver to provide name-to-address lookup functionality
that applications need for connecting to I2P destinations by name.

#### func  NewFullSAMResolver

```go
func NewFullSAMResolver(address string) (*SAMResolver, error)
```
NewFullSAMResolver creates a new complete SAM resolver by establishing its own
connection. This convenience function creates both a SAM connection and resolver
in a single operation. It's useful when you only need name resolution and don't
require a persistent SAM connection for session management or other operations.

The resolver will establish its own connection to the specified address and be
ready for immediate use. The caller is responsible for closing the resolver when
done.

Example:

    resolver, err := NewFullSAMResolver("127.0.0.1:7656")
    if err != nil {
        return err
    }
    defer resolver.Close()

#### func  NewSAMResolver

```go
func NewSAMResolver(parent *SAM) (*SAMResolver, error)
```
NewSAMResolver creates a new SAM resolver instance for I2P name lookups. This
function creates a resolver that can translate I2P names (like "example.i2p")
into Base32 destination addresses for use in connections and messaging.

The resolver uses the provided SAM connection for performing lookups through the
I2P network's address book and naming services. It's essential for applications
that want to connect to I2P services using human-readable names.

Example:

    resolver, err := NewSAMResolver(sam)
    if err != nil {
        return err
    }
    addr, err := resolver.Resolve("example.i2p")

#### type StreamListener

```go
type StreamListener = stream.StreamListener
```

StreamListener implements net.Listener for I2P streaming connections. It manages
incoming connection acceptance and provides thread-safe operations for accepting
connections from remote I2P destinations in server applications.

#### type StreamSession

```go
type StreamSession = stream.StreamSession
```

StreamSession provides TCP-like reliable connection capabilities over I2P
networks. It supports both client and server operations with connection
multiplexing, listener management, and standard Go networking interfaces for
streaming data.

#### type TestListener

```go
type TestListener struct {
}
```

TestListener manages a local I2P listener for testing purposes. It provides a
stable, local destination that can replace external sites in tests.

#### func  SetupTestListener

```go
func SetupTestListener(t *testing.T, config *TestListenerConfig) *TestListener
```
SetupTestListener creates and starts a local I2P listener that can serve as a
test destination. This replaces the need for external sites like i2p-projekt.i2p
or idk.i2p in tests. The listener will respond to HTTP GET requests with basic
HTML content.

#### func  SetupTestListenerWithHTTP

```go
func SetupTestListenerWithHTTP(t *testing.T, sessionID string) *TestListener
```
SetupTestListenerWithHTTP creates a test listener that provides HTTP-like
responses suitable for replacing external web sites in tests.

#### func (*TestListener) Addr

```go
func (tl *TestListener) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P address of the test listener.

#### func (*TestListener) AddrString

```go
func (tl *TestListener) AddrString() string
```
AddrString returns the Base32 address string of the test listener.

#### func (*TestListener) Close

```go
func (tl *TestListener) Close() error
```
Close shuts down the test listener and cleans up resources.

#### type TestListenerConfig

```go
type TestListenerConfig struct {
	SessionID    string
	HTTPResponse string // Optional custom HTTP response content
	Timeout      time.Duration
}
```

TestListenerConfig holds configuration for creating test listeners.

#### func  DefaultTestListenerConfig

```go
func DefaultTestListenerConfig(sessionID string) *TestListenerConfig
```
DefaultTestListenerConfig returns a default configuration for test listeners.
