# common
--
    import "github.com/go-i2p/go-sam-go/common"


## Usage

```go
const (
	DEFAULT_SAM_MIN = "3.1"
	DEFAULT_SAM_MAX = "3.3"
)
```

```go
const (
	SESSION_OK             = "SESSION STATUS RESULT=OK DESTINATION="
	SESSION_DUPLICATE_ID   = "SESSION STATUS RESULT=DUPLICATED_ID\n"
	SESSION_DUPLICATE_DEST = "SESSION STATUS RESULT=DUPLICATED_DEST\n"
	SESSION_INVALID_KEY    = "SESSION STATUS RESULT=INVALID_KEY\n"
	SESSION_I2P_ERROR      = "SESSION STATUS RESULT=I2P_ERROR MESSAGE="
)
```

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

```go
const (
	SAM_RESULT_OK            = "RESULT=OK"
	SAM_RESULT_INVALID_KEY   = "RESULT=INVALID_KEY"
	SAM_RESULT_KEY_NOT_FOUND = "RESULT=KEY_NOT_FOUND"
)
```

```go
const (
	HELLO_REPLY_OK        = "HELLO REPLY RESULT=OK"
	HELLO_REPLY_NOVERSION = "HELLO REPLY RESULT=NOVERSION\n"
)
```

```go
const (
	SESSION_STYLE_STREAM   = "STREAM"
	SESSION_STYLE_DATAGRAM = "DATAGRAM"
	SESSION_STYLE_RAW      = "RAW"
)
```

```go
const (
	ACCESS_TYPE_WHITELIST = "whitelist"
	ACCESS_TYPE_BLACKLIST = "blacklist"
	ACCESS_TYPE_NONE      = "none"
)
```

#### func  ExtractDest

```go
func ExtractDest(input string) string
```

#### func  ExtractPairInt

```go
func ExtractPairInt(input, value string) int
```

#### func  ExtractPairString

```go
func ExtractPairString(input, value string) string
```

#### func  IgnorePortError

```go
func IgnorePortError(err error) error
```

#### func  RandPort

```go
func RandPort() (portNumber string, err error)
```

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
SetFastRecieve tells clients to use compression

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

#### type BaseSession

```go
type BaseSession struct {
	SAM SAM
}
```


#### func (*BaseSession) Close

```go
func (bs *BaseSession) Close() error
```

#### func (*BaseSession) From

```go
func (bs *BaseSession) From() string
```

#### func (*BaseSession) ID

```go
func (bs *BaseSession) ID() string
```

#### func (*BaseSession) Keys

```go
func (bs *BaseSession) Keys() i2pkeys.I2PKeys
```

#### func (*BaseSession) LocalAddr

```go
func (bs *BaseSession) LocalAddr() net.Addr
```

#### func (*BaseSession) Read

```go
func (bs *BaseSession) Read(b []byte) (int, error)
```

#### func (*BaseSession) RemoteAddr

```go
func (bs *BaseSession) RemoteAddr() net.Addr
```

#### func (*BaseSession) SetDeadline

```go
func (bs *BaseSession) SetDeadline(t time.Time) error
```

#### func (*BaseSession) SetReadDeadline

```go
func (bs *BaseSession) SetReadDeadline(t time.Time) error
```

#### func (*BaseSession) SetWriteDeadline

```go
func (bs *BaseSession) SetWriteDeadline(t time.Time) error
```

#### func (*BaseSession) To

```go
func (bs *BaseSession) To() string
```

#### func (*BaseSession) Write

```go
func (bs *BaseSession) Write(b []byte) (int, error)
```

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

#### func (*I2PConfig) InboundLength

```go
func (f *I2PConfig) InboundLength() string
```

#### func (*I2PConfig) InboundLengthVariance

```go
func (f *I2PConfig) InboundLengthVariance() string
```

#### func (*I2PConfig) InboundQuantity

```go
func (f *I2PConfig) InboundQuantity() string
```

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

#### func (*I2PConfig) OutboundLength

```go
func (f *I2PConfig) OutboundLength() string
```

#### func (*I2PConfig) OutboundLengthVariance

```go
func (f *I2PConfig) OutboundLengthVariance() string
```

#### func (*I2PConfig) OutboundQuantity

```go
func (f *I2PConfig) OutboundQuantity() string
```

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
if keyfile fname does not exist

#### func (*SAM) Keys

```go
func (sam *SAM) Keys() (k *i2pkeys.I2PKeys)
```

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


#### func  NewEmit

```go
func NewEmit(opts ...func(*SAMEmit) error) (*SAMEmit, error)
```

#### func (*SAMEmit) Accept

```go
func (e *SAMEmit) Accept() string
```

#### func (*SAMEmit) AcceptBytes

```go
func (e *SAMEmit) AcceptBytes() []byte
```

#### func (*SAMEmit) Connect

```go
func (e *SAMEmit) Connect(dest string) string
```

#### func (*SAMEmit) ConnectBytes

```go
func (e *SAMEmit) ConnectBytes(dest string) []byte
```

#### func (*SAMEmit) Create

```go
func (e *SAMEmit) Create() string
```

#### func (*SAMEmit) CreateBytes

```go
func (e *SAMEmit) CreateBytes() []byte
```

#### func (*SAMEmit) GenerateDestination

```go
func (e *SAMEmit) GenerateDestination() string
```

#### func (*SAMEmit) GenerateDestinationBytes

```go
func (e *SAMEmit) GenerateDestinationBytes() []byte
```

#### func (*SAMEmit) Hello

```go
func (e *SAMEmit) Hello() string
```

#### func (*SAMEmit) HelloBytes

```go
func (e *SAMEmit) HelloBytes() []byte
```

#### func (*SAMEmit) Lookup

```go
func (e *SAMEmit) Lookup(name string) string
```

#### func (*SAMEmit) LookupBytes

```go
func (e *SAMEmit) LookupBytes(name string) []byte
```

#### func (*SAMEmit) SamOptionsString

```go
func (e *SAMEmit) SamOptionsString() string
```

#### type SAMResolver

```go
type SAMResolver struct {
	*SAM
}
```


#### func  NewFullSAMResolver

```go
func NewFullSAMResolver(address string) (*SAMResolver, error)
```

#### func  NewSAMResolver

```go
func NewSAMResolver(parent *SAM) (*SAMResolver, error)
```

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
