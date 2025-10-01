# common
--
    import "github.com/go-i2p/go-sam-go/common"


## Usage

```go
const (
	DEFAULT_SAM_MIN = "3.1"
	// DEFAULT_SAM_MAX specifies the maximum supported SAM protocol version.
	// This allows the library to work with newer SAM protocol features when available.
	DEFAULT_SAM_MAX = "3.3"
)
```
DEFAULT_SAM_MIN specifies the minimum supported SAM protocol version. This
constant is used during SAM bridge handshake to negotiate protocol
compatibility.

```go
const (
	SESSION_OK             = "SESSION STATUS RESULT=OK DESTINATION="
	SESSION_DUPLICATE_ID   = "SESSION STATUS RESULT=DUPLICATED_ID\n"
	SESSION_DUPLICATE_DEST = "SESSION STATUS RESULT=DUPLICATED_DEST\n"
	SESSION_INVALID_KEY    = "SESSION STATUS RESULT=INVALID_KEY\n"
	SESSION_I2P_ERROR      = "SESSION STATUS RESULT=I2P_ERROR MESSAGE="
)
```
SESSION_OK indicates successful session creation with destination key.
SESSION_DUPLICATE_ID indicates session creation failed due to duplicate session
ID. SESSION_DUPLICATE_DEST indicates session creation failed due to duplicate
destination. SESSION_INVALID_KEY indicates session creation failed due to
invalid destination key. SESSION_I2P_ERROR indicates session creation failed due
to I2P router error.

```go
const (
	SIG_NONE                 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"
	SIG_DSA_SHA1             = "SIGNATURE_TYPE=DSA_SHA1"
	SIG_ECDSA_SHA256_P256    = "SIGNATURE_TYPE=ECDSA_SHA256_P256"
	SIG_ECDSA_SHA384_P384    = "SIGNATURE_TYPE=ECDSA_SHA384_P384"
	SIG_ECDSA_SHA512_P521    = "SIGNATURE_TYPE=ECDSA_SHA512_P521"
	SIG_EdDSA_SHA512_Ed25519 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"
	// Add a default constant that points to the recommended secure signature type
	SIG_DEFAULT = SIG_EdDSA_SHA512_Ed25519
)
```
SIG_NONE is deprecated, use SIG_DEFAULT instead for secure signatures.
SIG_DSA_SHA1 specifies DSA with SHA1 signature type (legacy, not recommended).
SIG_ECDSA_SHA256_P256 specifies ECDSA with SHA256 on P256 curve signature type.
SIG_ECDSA_SHA384_P384 specifies ECDSA with SHA384 on P384 curve signature type.
SIG_ECDSA_SHA512_P521 specifies ECDSA with SHA512 on P521 curve signature type.
SIG_EdDSA_SHA512_Ed25519 specifies EdDSA with SHA512 on Ed25519 curve signature
type. SIG_DEFAULT points to the recommended secure signature type for new
applications.

```go
const (
	SAM_RESULT_OK            = "RESULT=OK"
	SAM_RESULT_INVALID_KEY   = "RESULT=INVALID_KEY"
	SAM_RESULT_KEY_NOT_FOUND = "RESULT=KEY_NOT_FOUND"
)
```
SAM_RESULT_OK indicates successful SAM operation completion.
SAM_RESULT_INVALID_KEY indicates SAM operation failed due to invalid key format.
SAM_RESULT_KEY_NOT_FOUND indicates SAM operation failed due to missing key.

```go
const (
	HELLO_REPLY_OK        = "HELLO REPLY RESULT=OK"
	HELLO_REPLY_NOVERSION = "HELLO REPLY RESULT=NOVERSION\n"
)
```
HELLO_REPLY_OK indicates successful SAM handshake completion.
HELLO_REPLY_NOVERSION indicates SAM handshake failed due to unsupported protocol
version.

```go
const (
	SESSION_STYLE_STREAM   = "STREAM"
	SESSION_STYLE_DATAGRAM = "DATAGRAM"
	SESSION_STYLE_RAW      = "RAW"
)
```
SESSION_STYLE_STREAM creates TCP-like reliable connection sessions.
SESSION_STYLE_DATAGRAM creates UDP-like message-based sessions.
SESSION_STYLE_RAW creates low-level packet transmission sessions.

```go
const (
	ACCESS_TYPE_WHITELIST = "whitelist"
	ACCESS_TYPE_BLACKLIST = "blacklist"
	ACCESS_TYPE_NONE      = "none"
)
```
ACCESS_TYPE_WHITELIST allows only specified destinations in access list.
ACCESS_TYPE_BLACKLIST blocks specified destinations in access list.
ACCESS_TYPE_NONE disables access list filtering entirely.

#### func  ExtractDest

```go
func ExtractDest(input string) string
```
ExtractDest extracts the destination address from a SAM protocol response. Takes
the first space-separated token from the input string as the destination. Used
for parsing SAM session creation responses and connection messages.

#### func  ExtractPairInt

```go
func ExtractPairInt(input, value string) int
```
ExtractPairInt extracts an integer value from a key=value pair in a
space-separated string. Uses ExtractPairString internally and converts the
result to integer. Returns 0 if the key is not found or the value cannot be
converted to integer.

#### func  ExtractPairString

```go
func ExtractPairString(input, value string) string
```
ExtractPairString extracts the value from a key=value pair in a space-separated
string. Searches for the specified key prefix and returns the associated value.
Returns empty string if the key is not found or has no value.

#### func  IgnorePortError

```go
func IgnorePortError(err error) error
```
IgnorePortError filters out "missing port in address" errors for convenience.
This function is used when parsing addresses that may not include port numbers.
Returns nil if the error is about missing port, otherwise returns the original
error.

#### func  RandPort

```go
func RandPort() (portNumber string, err error)
```
RandPort generates a random available port number for local testing. Attempts to
find a port that is available for both TCP and UDP connections. Returns the port
as a string or an error if no available port is found after 30 attempts.

#### func  SetAccessList

```go
func SetAccessList(s []string) func(*SAMEmit) error
```
SetAccessList tells the system to treat the AccessList as a whitelist

#### func  SetAccessListType

```go
func SetAccessListType(s string) func(*SAMEmit) error
```
SetAccessListType tells the system to treat the AccessList as a whitelist

#### func  SetAllowZeroIn

```go
func SetAllowZeroIn(b bool) func(*SAMEmit) error
```
SetAllowZeroIn tells the tunnel to accept zero-hop peers

#### func  SetAllowZeroOut

```go
func SetAllowZeroOut(b bool) func(*SAMEmit) error
```
SetAllowZeroOut tells the tunnel to accept zero-hop peers

#### func  SetCloseIdle

```go
func SetCloseIdle(b bool) func(*SAMEmit) error
```
SetCloseIdle tells the connection to close it's tunnels during extended idle
time.

#### func  SetCloseIdleTime

```go
func SetCloseIdleTime(u int) func(*SAMEmit) error
```
SetCloseIdleTime sets the time to wait before closing tunnels to idle levels

#### func  SetCloseIdleTimeMs

```go
func SetCloseIdleTimeMs(u int) func(*SAMEmit) error
```
SetCloseIdleTimeMs sets the time to wait before closing tunnels to idle levels
in milliseconds

#### func  SetCompress

```go
func SetCompress(b bool) func(*SAMEmit) error
```
SetCompress tells clients to use compression

#### func  SetEncrypt

```go
func SetEncrypt(b bool) func(*SAMEmit) error
```
SetEncrypt tells the router to use an encrypted leaseset

#### func  SetFastRecieve

```go
func SetFastRecieve(b bool) func(*SAMEmit) error
```
SetFastRecieve enables or disables fast receive mode for improved performance.
When enabled, allows bypassing some protocol overhead for faster data
transmission. SetFastRecieve tells clients to use compression

#### func  SetInBackups

```go
func SetInBackups(u int) func(*SAMEmit) error
```
SetInBackups sets the inbound tunnel backups

#### func  SetInLength

```go
func SetInLength(u int) func(*SAMEmit) error
```
SetInLength sets the number of hops inbound

#### func  SetInQuantity

```go
func SetInQuantity(u int) func(*SAMEmit) error
```
SetInQuantity sets the inbound tunnel quantity

#### func  SetInVariance

```go
func SetInVariance(i int) func(*SAMEmit) error
```
SetInVariance sets the variance of a number of hops inbound

#### func  SetLeaseSetKey

```go
func SetLeaseSetKey(s string) func(*SAMEmit) error
```
SetLeaseSetKey sets the host of the SAMEmit's SAM bridge

#### func  SetLeaseSetPrivateKey

```go
func SetLeaseSetPrivateKey(s string) func(*SAMEmit) error
```
SetLeaseSetPrivateKey sets the host of the SAMEmit's SAM bridge

#### func  SetLeaseSetPrivateSigningKey

```go
func SetLeaseSetPrivateSigningKey(s string) func(*SAMEmit) error
```
SetLeaseSetPrivateSigningKey sets the host of the SAMEmit's SAM bridge

#### func  SetMessageReliability

```go
func SetMessageReliability(s string) func(*SAMEmit) error
```
SetMessageReliability sets the host of the SAMEmit's SAM bridge

#### func  SetName

```go
func SetName(s string) func(*SAMEmit) error
```
SetName sets the host of the SAMEmit's SAM bridge

#### func  SetOutBackups

```go
func SetOutBackups(u int) func(*SAMEmit) error
```
SetOutBackups sets the inbound tunnel backups

#### func  SetOutLength

```go
func SetOutLength(u int) func(*SAMEmit) error
```
SetOutLength sets the number of hops outbound

#### func  SetOutQuantity

```go
func SetOutQuantity(u int) func(*SAMEmit) error
```
SetOutQuantity sets the outbound tunnel quantity

#### func  SetOutVariance

```go
func SetOutVariance(i int) func(*SAMEmit) error
```
SetOutVariance sets the variance of a number of hops outbound

#### func  SetReduceIdle

```go
func SetReduceIdle(b bool) func(*SAMEmit) error
```
SetReduceIdle tells the connection to reduce it's tunnels during extended idle
time.

#### func  SetReduceIdleQuantity

```go
func SetReduceIdleQuantity(u int) func(*SAMEmit) error
```
SetReduceIdleQuantity sets minimum number of tunnels to reduce to during idle
time

#### func  SetReduceIdleTime

```go
func SetReduceIdleTime(u int) func(*SAMEmit) error
```
SetReduceIdleTime sets the time to wait before reducing tunnels to idle levels

#### func  SetReduceIdleTimeMs

```go
func SetReduceIdleTimeMs(u int) func(*SAMEmit) error
```
SetReduceIdleTimeMs sets the time to wait before reducing tunnels to idle levels
in milliseconds

#### func  SetSAMAddress

```go
func SetSAMAddress(s string) func(*SAMEmit) error
```
SetSAMAddress sets the SAM address all-at-once

#### func  SetSAMHost

```go
func SetSAMHost(s string) func(*SAMEmit) error
```
SetSAMHost sets the host of the SAMEmit's SAM bridge

#### func  SetSAMPort

```go
func SetSAMPort(s string) func(*SAMEmit) error
```
SetSAMPort sets the port of the SAMEmit's SAM bridge using a string

#### func  SetType

```go
func SetType(s string) func(*SAMEmit) error
```
SetType sets the type of the forwarder server

#### func  SplitHostPort

```go
func SplitHostPort(hostport string) (string, string, error)
```
SplitHostPort separates host and port from a combined address string. Unlike
net.SplitHostPort, this function handles addresses without ports gracefully.
Returns host, port as strings, and error. Port defaults to "0" if not specified.

#### type BaseSession

```go
type BaseSession struct {
	SAM SAM
}
```

BaseSession provides the underlying SAM session functionality. It manages the
connection to the SAM bridge and handles session lifecycle.

#### func (*BaseSession) Close

```go
func (bs *BaseSession) Close() error
```
Close closes the session connection and releases associated resources.
Implements the io.Closer interface for proper resource cleanup.

#### func (*BaseSession) Conn

```go
func (bs *BaseSession) Conn() net.Conn
```
Conn returns the underlying network connection for the session. This provides
access to the raw connection for advanced operations.

#### func (*BaseSession) From

```go
func (bs *BaseSession) From() string
```
From returns the configured source port for the session. Used in port-based
session configurations for service identification.

#### func (*BaseSession) ID

```go
func (bs *BaseSession) ID() string
```
ID returns the unique session identifier used by the SAM bridge. This identifier
is used to distinguish between multiple sessions on the same connection.

#### func (*BaseSession) Keys

```go
func (bs *BaseSession) Keys() i2pkeys.I2PKeys
```
Keys returns the I2P cryptographic keys associated with this session. These keys
define the session's I2P destination and identity.

#### func (*BaseSession) LocalAddr

```go
func (bs *BaseSession) LocalAddr() net.Addr
```
LocalAddr returns the local network address of the session connection.
Implements the net.Conn interface for network address information.

#### func (*BaseSession) Read

```go
func (bs *BaseSession) Read(b []byte) (int, error)
```
Read reads data from the session connection into the provided buffer. Implements
the io.Reader interface for standard Go I/O operations.

#### func (*BaseSession) RemoteAddr

```go
func (bs *BaseSession) RemoteAddr() net.Addr
```
RemoteAddr returns the remote network address of the session connection.
Implements the net.Conn interface for network address information.

#### func (*BaseSession) SetDeadline

```go
func (bs *BaseSession) SetDeadline(t time.Time) error
```
SetDeadline sets read and write deadlines for the session connection. Implements
the net.Conn interface for timeout control.

#### func (*BaseSession) SetReadDeadline

```go
func (bs *BaseSession) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets the read deadline for the session connection. Implements
the net.Conn interface for read timeout control.

#### func (*BaseSession) SetWriteDeadline

```go
func (bs *BaseSession) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets the write deadline for the session connection. Implements
the net.Conn interface for write timeout control.

#### func (*BaseSession) To

```go
func (bs *BaseSession) To() string
```
To returns the configured destination port for the session. Used in port-based
session configurations for service identification.

#### func (*BaseSession) Write

```go
func (bs *BaseSession) Write(b []byte) (int, error)
```
Write writes data from the buffer to the session connection. Implements the
io.Writer interface for standard Go I/O operations.

#### type I2PConfig

```go
type I2PConfig struct {
	SamHost string
	SamPort int
	TunName string

	SamMin string
	SamMax string

	Fromport string
	Toport   string

	Style   string
	TunType string

	DestinationKeys *i2pkeys.I2PKeys

	SigType                   string
	EncryptLeaseSet           bool
	LeaseSetKey               string
	LeaseSetPrivateKey        string
	LeaseSetPrivateSigningKey string
	LeaseSetKeys              i2pkeys.I2PKeys
	InAllowZeroHop            bool
	OutAllowZeroHop           bool
	InLength                  int
	OutLength                 int
	InQuantity                int
	OutQuantity               int
	InVariance                int
	OutVariance               int
	InBackupQuantity          int
	OutBackupQuantity         int
	FastRecieve               bool
	UseCompression            bool
	MessageReliability        string
	CloseIdle                 bool
	CloseIdleTime             int
	ReduceIdle                bool
	ReduceIdleTime            int
	ReduceIdleQuantity        int
	LeaseSetEncryption        string

	// Streaming Library options
	AccessListType string
	AccessList     []string
}
```

I2PConfig is a struct which manages I2P configuration options.

#### func  NewConfig

```go
func NewConfig(opts ...func(*I2PConfig) error) (*I2PConfig, error)
```
NewConfig creates a new I2PConfig instance with default values and applies
functional options. Returns a configured instance ready for use in session
creation or an error if any option fails. Example: config, err :=
NewConfig(SetInLength(4), SetOutLength(4))

#### func (*I2PConfig) Accesslist

```go
func (f *I2PConfig) Accesslist() string
```
Accesslist generates the I2CP access list configuration string based on the
configured access list

#### func (*I2PConfig) Accesslisttype

```go
func (f *I2PConfig) Accesslisttype() string
```
Accesslisttype returns the I2CP access list configuration string based on the
AccessListType setting

#### func (*I2PConfig) Close

```go
func (f *I2PConfig) Close() string
```
Close returns I2CP close-on-idle configuration settings as a string if enabled

#### func (*I2PConfig) DestinationKey

```go
func (f *I2PConfig) DestinationKey() string
```
DestinationKey returns the DESTINATION configuration string for the SAM bridge
If destination keys are set, returns them as a string, otherwise returns
"TRANSIENT"

#### func (*I2PConfig) DoZero

```go
func (f *I2PConfig) DoZero() string
```
DoZero returns the zero hop and fast receive configuration string settings

#### func (*I2PConfig) EncryptLease

```go
func (f *I2PConfig) EncryptLease() string
```
EncryptLease returns the lease set encryption configuration string Returns
"i2cp.encryptLeaseSet=true" if encryption is enabled, empty string otherwise

#### func (*I2PConfig) FromPort

```go
func (f *I2PConfig) FromPort() string
```
FromPort returns the FROM_PORT configuration string for SAM bridges >= 3.1
Returns an empty string if SAM version < 3.1 or if fromport is "0"

#### func (*I2PConfig) ID

```go
func (f *I2PConfig) ID() string
```
ID returns the tunnel name as a formatted string. If no tunnel name is set,
generates a random 12-character name using lowercase letters.

#### func (*I2PConfig) InboundBackupQuantity

```go
func (f *I2PConfig) InboundBackupQuantity() string
```
InboundBackupQuantity returns the inbound tunnel backup quantity configuration
string. Specifies the number of backup tunnels to maintain for inbound
connections.

#### func (*I2PConfig) InboundLength

```go
func (f *I2PConfig) InboundLength() string
```
InboundLength returns the inbound tunnel length configuration string. Specifies
the desired length of the inbound tunnel (number of hops).

#### func (*I2PConfig) InboundLengthVariance

```go
func (f *I2PConfig) InboundLengthVariance() string
```
InboundLengthVariance returns the inbound tunnel length variance configuration
string. Controls the randomness in inbound tunnel hop count for improved
anonymity.

#### func (*I2PConfig) InboundQuantity

```go
func (f *I2PConfig) InboundQuantity() string
```
InboundQuantity returns the inbound tunnel quantity configuration string.
Specifies the number of parallel inbound tunnels to maintain for load balancing.

#### func (*I2PConfig) LeaseSetEncryptionType

```go
func (f *I2PConfig) LeaseSetEncryptionType() string
```
LeaseSetEncryptionType returns the I2CP lease set encryption type configuration
string. If no encryption type is set, returns default value "4,0". Validates
that all encryption types are valid integers.

#### func (*I2PConfig) LeaseSetSettings

```go
func (f *I2PConfig) LeaseSetSettings() (string, string, string)
```
Leasesetsettings returns the lease set configuration strings for I2P Returns
three strings: lease set key, private key, and private signing key settings

#### func (*I2PConfig) MaxSAM

```go
func (f *I2PConfig) MaxSAM() string
```
MaxSAM returns the maximum SAM version supported as a string If no maximum
version is set, returns default value "3.1"

#### func (*I2PConfig) MinSAM

```go
func (f *I2PConfig) MinSAM() string
```
MinSAM returns the minimum SAM version supported as a string If no minimum
version is set, returns default value "3.0"

#### func (*I2PConfig) OutboundBackupQuantity

```go
func (f *I2PConfig) OutboundBackupQuantity() string
```
OutboundBackupQuantity returns the outbound tunnel backup quantity configuration
string. Specifies the number of backup tunnels to maintain for outbound
connections.

#### func (*I2PConfig) OutboundLength

```go
func (f *I2PConfig) OutboundLength() string
```
OutboundLength returns the outbound tunnel length configuration string.
Specifies the desired length of the outbound tunnel (number of hops).

#### func (*I2PConfig) OutboundLengthVariance

```go
func (f *I2PConfig) OutboundLengthVariance() string
```
OutboundLengthVariance returns the outbound tunnel length variance configuration
string. Controls the randomness in outbound tunnel hop count for improved
anonymity.

#### func (*I2PConfig) OutboundQuantity

```go
func (f *I2PConfig) OutboundQuantity() string
```
OutboundQuantity returns the outbound tunnel quantity configuration string.
Specifies the number of parallel outbound tunnels to maintain for load
balancing.

#### func (*I2PConfig) Print

```go
func (f *I2PConfig) Print() []string
```
Print returns a slice of strings containing all the I2P configuration settings

#### func (*I2PConfig) Reduce

```go
func (f *I2PConfig) Reduce() string
```
Reduce returns I2CP reduce-on-idle configuration settings as a string if enabled

#### func (*I2PConfig) Reliability

```go
func (f *I2PConfig) Reliability() string
```
Reliability returns the message reliability configuration string for the SAM
bridge If a reliability setting is specified, returns formatted
i2cp.messageReliability setting

#### func (*I2PConfig) SAMAddress

```go
func (f *I2PConfig) SAMAddress() string
```
SAMAddress returns the SAM bridge address in the format "host:port" This is a
convenience method that uses the Sam() function to get the address. It is used
to provide a consistent interface for retrieving the SAM address.

#### func (*I2PConfig) Sam

```go
func (f *I2PConfig) Sam() string
```
Sam returns the SAM bridge address as a string in the format "host:port"

#### func (*I2PConfig) SessionStyle

```go
func (f *I2PConfig) SessionStyle() string
```
SessionStyle returns the SAM session style configuration string If no style is
set, defaults to "STREAM"

#### func (*I2PConfig) SetSAMAddress

```go
func (f *I2PConfig) SetSAMAddress(addr string)
```
SetSAMAddress sets the SAM bridge host and port from a combined address string.
If no address is provided, it sets default values for the host and port.

#### func (*I2PConfig) SignatureType

```go
func (f *I2PConfig) SignatureType() string
```
SignatureType returns the SIGNATURE_TYPE configuration string for SAM bridges >=
3.1 Returns empty string if SAM version < 3.1 or if no signature type is set

#### func (*I2PConfig) ToPort

```go
func (f *I2PConfig) ToPort() string
```
ToPort returns the TO_PORT configuration string for SAM bridges >= 3.1 Returns
an empty string if SAM version < 3.1 or if toport is "0"

#### func (*I2PConfig) UsingCompression

```go
func (f *I2PConfig) UsingCompression() string
```
UsingCompression returns the compression configuration string for I2P streams.
Enables or disables data compression to reduce bandwidth usage at the cost of
CPU overhead.

#### type Option

```go
type Option func(*SAMEmit) error
```

Option is a SAMEmit Option

#### type Options

```go
type Options map[string]string
```

options map

#### func (Options) AsList

```go
func (opts Options) AsList() (ls []string)
```
obtain sam options as list of strings

#### type SAM

```go
type SAM struct {
	SAMEmit
	SAMResolver
	net.Conn

	// Timeout for SAM connections
	Timeout time.Duration
	// Context for control of lifecycle
	Context context.Context
}
```

Used for controlling I2Ps SAMv3.

#### func  NewSAM

```go
func NewSAM(address string) (*SAM, error)
```
NewSAM creates a new SAM instance by connecting to the specified address,
performing the hello handshake, and initializing the SAM resolver. It returns a
pointer to the SAM instance or an error if any step fails. This function
combines connection establishment and hello handshake into a single step,
eliminating the need for separate helper functions. It also initializes the SAM
resolver directly after the connection is established. The SAM instance is ready
to use for further operations like session creation or name resolution.

#### func (*SAM) Close

```go
func (sam *SAM) Close() error
```
close this sam session

#### func (*SAM) EnsureKeyfile

```go
func (sam *SAM) EnsureKeyfile(fname string) (keys i2pkeys.I2PKeys, err error)
```
EnsureKeyfile ensures cryptographic keys are available, either by generating
transient keys or by loading/creating persistent keys from the specified file.

#### func (*SAM) Keys

```go
func (sam *SAM) Keys() (k *i2pkeys.I2PKeys)
```
Keys retrieves the I2P destination keys associated with this SAM instance.
Returns a pointer to the keys used for this SAM session's I2P identity.

#### func (*SAM) Lookup

```go
func (sam *SAM) Lookup(name string) (i2pkeys.I2PAddr, error)
```
Performs a lookup, probably this order: 1) routers known addresses, cached
addresses, 3) by asking peers in the I2P network.

#### func (SAM) NewGenericSession

```go
func (sam SAM) NewGenericSession(style, id string, keys i2pkeys.I2PKeys, extras []string) (Session, error)
```
Creates a new session with the style of either "STREAM", "DATAGRAM" or "RAW",
for a new I2P tunnel with name id, using the cypher keys specified, with the
I2CP/streaminglib-options as specified. Extra arguments can be specified by
setting extra to something else than []string{}. This sam3 instance is now a
session

#### func (SAM) NewGenericSessionWithSignature

```go
func (sam SAM) NewGenericSessionWithSignature(style, id string, keys i2pkeys.I2PKeys, sigType string, extras []string) (Session, error)
```
Creates a new session with the style of either "STREAM", "DATAGRAM" or "RAW",
for a new I2P tunnel with name id, using the cypher keys specified, with the
I2CP/streaminglib-options as specified. Extra arguments can be specified by
setting extra to something else than []string{}. This sam3 instance is now a
session

#### func (SAM) NewGenericSessionWithSignatureAndPorts

```go
func (sam SAM) NewGenericSessionWithSignatureAndPorts(style, id, from, to string, keys i2pkeys.I2PKeys, sigType string, extras []string) (Session, error)
```
Creates a new session with the style of either "STREAM", "DATAGRAM" or "RAW",
for a new I2P tunnel with name id, using the cypher keys specified, with the
I2CP/streaminglib-options as specified. Extra arguments can be specified by
setting extra to something else than []string{}. This sam3 instance is now a
session

#### func (*SAM) NewKeys

```go
func (sam *SAM) NewKeys(sigType ...string) (i2pkeys.I2PKeys, error)
```
Creates the I2P-equivalent of an IP address, that is unique and only the one who
has the private keys can send messages from. The public keys are the I2P
desination (the address) that anyone can send messages to.

#### func (*SAM) ReadKeys

```go
func (sam *SAM) ReadKeys(r io.Reader) (err error)
```
read public/private keys from an io.Reader

#### type SAMEmit

```go
type SAMEmit struct {
	I2PConfig
}
```

SAMEmit handles SAM protocol message generation and configuration. It embeds
I2PConfig to provide access to all tunnel and session configuration options.

#### func  NewEmit

```go
func NewEmit(opts ...func(*SAMEmit) error) (*SAMEmit, error)
```
NewEmit creates a new SAMEmit instance with the specified configuration options.
Applies functional options to configure the emitter with custom settings.
Returns an error if any option fails to apply correctly.

#### func (*SAMEmit) Accept

```go
func (e *SAMEmit) Accept() string
```
Accept generates a SAM STREAM ACCEPT command for accepting incoming connections.
Creates a command to listen for and accept connections on the configured
session.

#### func (*SAMEmit) AcceptBytes

```go
func (e *SAMEmit) AcceptBytes() []byte
```
AcceptBytes returns the STREAM ACCEPT command as bytes for transmission.
Convenience method for sending accept requests over network connections.

#### func (*SAMEmit) Connect

```go
func (e *SAMEmit) Connect(dest string) string
```
Connect generates a SAM STREAM CONNECT command for establishing connections.
Takes a destination address and creates a command to connect to that I2P
destination.

#### func (*SAMEmit) ConnectBytes

```go
func (e *SAMEmit) ConnectBytes(dest string) []byte
```
ConnectBytes returns the STREAM CONNECT command as bytes for transmission.
Convenience method for sending connection requests over network connections.

#### func (*SAMEmit) Create

```go
func (e *SAMEmit) Create() string
```
Create generates a SAM SESSION CREATE command for establishing new sessions.
Combines session style, ports, ID, destination, signature type, and options into
a single command.

#### func (*SAMEmit) CreateBytes

```go
func (e *SAMEmit) CreateBytes() []byte
```
CreateBytes returns the SESSION CREATE command as bytes for network
transmission. Includes debug output of the command for troubleshooting session
creation issues.

#### func (*SAMEmit) GenerateDestination

```go
func (e *SAMEmit) GenerateDestination() string
```
GenerateDestination creates a SAM DEST GENERATE command for key generation. Uses
the configured signature type to request new I2P destination keys from the
router.

#### func (*SAMEmit) GenerateDestinationBytes

```go
func (e *SAMEmit) GenerateDestinationBytes() []byte
```
GenerateDestinationBytes returns the DEST GENERATE command as bytes. Convenience
method for network transmission of key generation requests.

#### func (*SAMEmit) Hello

```go
func (e *SAMEmit) Hello() string
```
Hello generates the SAM protocol HELLO command for initial handshake. Includes
minimum and maximum supported SAM protocol versions for negotiation.

#### func (*SAMEmit) HelloBytes

```go
func (e *SAMEmit) HelloBytes() []byte
```
HelloBytes returns the HELLO command as a byte slice for network transmission.
Convenience method for sending the handshake command over network connections.

#### func (*SAMEmit) Lookup

```go
func (e *SAMEmit) Lookup(name string) string
```
Lookup generates a SAM NAMING LOOKUP command for address resolution. Takes a
human-readable name and creates a command to resolve it to an I2P destination.

#### func (*SAMEmit) LookupBytes

```go
func (e *SAMEmit) LookupBytes(name string) []byte
```
LookupBytes returns the NAMING LOOKUP command as bytes for transmission.
Convenience method for sending address resolution requests over network
connections.

#### func (*SAMEmit) SamOptionsString

```go
func (e *SAMEmit) SamOptionsString() string
```
SamOptionsString generates a space-separated string of all I2P configuration
options. Used internally to construct SAM protocol messages with tunnel and
session parameters.

#### type SAMResolver

```go
type SAMResolver struct {
	*SAM
}
```

SAMResolver provides I2P address resolution services through SAM protocol. It
maintains a connection to the SAM bridge for performing address lookups.

#### func  NewFullSAMResolver

```go
func NewFullSAMResolver(address string) (*SAMResolver, error)
```
NewFullSAMResolver creates a complete SAMResolver with its own SAM connection.
Establishes a new connection to the specified SAM bridge address for address
resolution. Returns a fully configured resolver or an error if connection fails.

#### func  NewSAMResolver

```go
func NewSAMResolver(parent *SAM) (*SAMResolver, error)
```
NewSAMResolver creates a new SAMResolver using an existing SAM instance. This
allows sharing a single SAM connection for both session management and address
resolution. Returns a configured resolver ready for performing I2P address
lookups.

#### func (*SAMResolver) Resolve

```go
func (sam *SAMResolver) Resolve(name string) (i2pkeys.I2PAddr, error)
```
Performs a lookup, probably this order: 1) routers known addresses, cached
addresses, 3) by asking peers in the I2P network.

#### type Session

```go
type Session interface {
	net.Conn
	ID() string
	Keys() i2pkeys.I2PKeys
	Close() error
}
```

Session represents a generic I2P session interface for different connection
types. It extends net.Conn with I2P-specific functionality for session
identification and key management. All session implementations (stream,
datagram, raw) must implement this interface.
