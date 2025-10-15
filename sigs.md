# sam3
--
    import "github.com/go-i2p/sam3"

Library for I2Ps SAMv3 bridge (https://geti2p.com)

## Usage

```go
const (
	Sig_NONE                 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"
	Sig_DSA_SHA1             = "SIGNATURE_TYPE=DSA_SHA1"
	Sig_ECDSA_SHA256_P256    = "SIGNATURE_TYPE=ECDSA_SHA256_P256"
	Sig_ECDSA_SHA384_P384    = "SIGNATURE_TYPE=ECDSA_SHA384_P384"
	Sig_ECDSA_SHA512_P521    = "SIGNATURE_TYPE=ECDSA_SHA512_P521"
	Sig_EdDSA_SHA512_Ed25519 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"
)
```

```go
var (
	// Suitable options if you are shuffling A LOT of traffic. If unused, this
	// will waste your resources.
	Options_Humongous = []string{"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=3", "outbound.backupQuantity=3",
		"inbound.quantity=6", "outbound.quantity=6"}

	// Suitable for shuffling a lot of traffic.
	Options_Large = []string{"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=1", "outbound.backupQuantity=1",
		"inbound.quantity=4", "outbound.quantity=4"}

	// Suitable for shuffling a lot of traffic quickly with minimum
	// anonymity. Uses 1 hop and multiple tunnels.
	Options_Wide = []string{"inbound.length=1", "outbound.length=1",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=2", "outbound.backupQuantity=2",
		"inbound.quantity=3", "outbound.quantity=3"}

	// Suitable for shuffling medium amounts of traffic.
	Options_Medium = []string{"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=2", "outbound.quantity=2"}

	// Sensible defaults for most people
	Options_Default = []string{"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=0", "outbound.lengthVariance=0",
		"inbound.backupQuantity=1", "outbound.backupQuantity=1",
		"inbound.quantity=1", "outbound.quantity=1"}

	// Suitable only for small dataflows, and very short lasting connections:
	// You only have one tunnel in each direction, so if any of the nodes
	// through which any of your two tunnels pass through go offline, there will
	// be a complete halt in the dataflow, until a new tunnel is built.
	Options_Small = []string{"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=1", "outbound.quantity=1"}

	// Does not use any anonymization, you connect directly to others tunnel
	// endpoints, thus revealing your identity but not theirs. Use this only
	// if you don't care.
	Options_Warning_ZeroHop = []string{"inbound.length=0", "outbound.length=0",
		"inbound.lengthVariance=0", "outbound.lengthVariance=0",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=2", "outbound.quantity=2"}
)
```
Examples and suggestions for options when creating sessions.

```go
var PrimarySessionSwitch string = PrimarySessionString()
```

```go
var SAM_HOST = getEnv("sam_host", "127.0.0.1")
```

```go
var SAM_PORT = getEnv("sam_port", "7656")
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

#### func  GenerateOptionString

```go
func GenerateOptionString(opts []string) string
```

#### func  GetSAM3Logger

```go
func GetSAM3Logger() *logrus.Logger
```
GetSAM3Logger returns the initialized logger

#### func  IgnorePortError

```go
func IgnorePortError(err error) error
```

#### func  InitializeSAM3Logger

```go
func InitializeSAM3Logger()
```

#### func  PrimarySessionString

```go
func PrimarySessionString() string
```

#### func  RandString

```go
func RandString() string
```

#### func  SAMDefaultAddr

```go
func SAMDefaultAddr(fallforward string) string
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

#### type DatagramSession

```go
type DatagramSession struct {
}
```

The DatagramSession implements net.PacketConn. It works almost like ordinary
UDP, except that datagrams may be at most 31kB large. These datagrams are also
end-to-end encrypted, signed and includes replay-protection. And they are also
built to be surveillance-resistant (yey!).

#### func (*DatagramSession) Accept

```go
func (s *DatagramSession) Accept() (net.Conn, error)
```

#### func (*DatagramSession) Addr

```go
func (s *DatagramSession) Addr() net.Addr
```

#### func (*DatagramSession) B32

```go
func (s *DatagramSession) B32() string
```

#### func (*DatagramSession) Close

```go
func (s *DatagramSession) Close() error
```
Closes the DatagramSession. Implements net.PacketConn

#### func (*DatagramSession) Dial

```go
func (s *DatagramSession) Dial(net string, addr string) (*DatagramSession, error)
```

#### func (*DatagramSession) DialI2PRemote

```go
func (s *DatagramSession) DialI2PRemote(net string, addr net.Addr) (*DatagramSession, error)
```

#### func (*DatagramSession) DialRemote

```go
func (s *DatagramSession) DialRemote(net, addr string) (net.PacketConn, error)
```

#### func (*DatagramSession) LocalAddr

```go
func (s *DatagramSession) LocalAddr() net.Addr
```
Implements net.PacketConn

#### func (*DatagramSession) LocalI2PAddr

```go
func (s *DatagramSession) LocalI2PAddr() i2pkeys.I2PAddr
```
Returns the I2P destination of the DatagramSession.

#### func (*DatagramSession) Lookup

```go
func (s *DatagramSession) Lookup(name string) (a net.Addr, err error)
```

#### func (*DatagramSession) Read

```go
func (s *DatagramSession) Read(b []byte) (n int, err error)
```

#### func (*DatagramSession) ReadFrom

```go
func (s *DatagramSession) ReadFrom(b []byte) (n int, addr net.Addr, err error)
```
Reads one datagram sent to the destination of the DatagramSession. Returns the
number of bytes read, from what address it was sent, or an error. implements
net.PacketConn

#### func (*DatagramSession) RemoteAddr

```go
func (s *DatagramSession) RemoteAddr() net.Addr
```

#### func (*DatagramSession) SetDeadline

```go
func (s *DatagramSession) SetDeadline(t time.Time) error
```
Sets read and write deadlines for the DatagramSession. Implements net.PacketConn
and does the same thing. Setting write deadlines for datagrams is seldom done.

#### func (*DatagramSession) SetReadDeadline

```go
func (s *DatagramSession) SetReadDeadline(t time.Time) error
```
Sets read deadline for the DatagramSession. Implements net.PacketConn

#### func (*DatagramSession) SetWriteBuffer

```go
func (s *DatagramSession) SetWriteBuffer(bytes int) error
```

#### func (*DatagramSession) SetWriteDeadline

```go
func (s *DatagramSession) SetWriteDeadline(t time.Time) error
```
Sets the write deadline for the DatagramSession. Implements net.Packetconn.

#### func (*DatagramSession) Write

```go
func (s *DatagramSession) Write(b []byte) (int, error)
```

#### func (*DatagramSession) WriteTo

```go
func (s *DatagramSession) WriteTo(b []byte, addr net.Addr) (n int, err error)
```
Sends one signed datagram to the destination specified. At the time of writing,
maximum size is 31 kilobyte, but this may change in the future. Implements
net.PacketConn.

#### type I2PConfig

```go
type I2PConfig struct {
	SamHost string
	SamPort string
	TunName string

	SamMin string
	SamMax string

	Fromport string
	Toport   string

	Style   string
	TunType string

	DestinationKeys i2pkeys.I2PKeys

	SigType                   string
	EncryptLeaseSet           string
	LeaseSetKey               string
	LeaseSetPrivateKey        string
	LeaseSetPrivateSigningKey string
	LeaseSetKeys              i2pkeys.I2PKeys
	InAllowZeroHop            string
	OutAllowZeroHop           string
	InLength                  string
	OutLength                 string
	InQuantity                string
	OutQuantity               string
	InVariance                string
	OutVariance               string
	InBackupQuantity          string
	OutBackupQuantity         string
	FastRecieve               string
	UseCompression            string
	MessageReliability        string
	CloseIdle                 string
	CloseIdleTime             string
	ReduceIdle                string
	ReduceIdleTime            string
	ReduceIdleQuantity        string
	LeaseSetEncryption        string

	//Streaming Library options
	AccessListType string
	AccessList     []string
}
```

I2PConfig is a struct which manages I2P configuration options

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

#### func (*I2PConfig) LeaseSetEncryptionType

```go
func (f *I2PConfig) LeaseSetEncryptionType() string
```
LeaseSetEncryptionType returns the I2CP lease set encryption type configuration
string. If no encryption type is set, returns default value "4,0". Validates
that all encryption types are valid integers.

#### func (*I2PConfig) Leasesetsettings

```go
func (f *I2PConfig) Leasesetsettings() (string, string, string)
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
SetSAMAddress sets the SAM bridge host and port from a combined address string
addr format can be either "host" or "host:port"

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

#### type PrimarySession

```go
type PrimarySession struct {
	Timeout  time.Duration
	Deadline time.Time

	Config SAMEmit
}
```

Represents a primary session.

#### func (*PrimarySession) Addr

```go
func (ss *PrimarySession) Addr() i2pkeys.I2PAddr
```
Returns the I2P destination (the address) of the stream session

#### func (*PrimarySession) Close

```go
func (ss *PrimarySession) Close() error
```

#### func (*PrimarySession) Dial

```go
func (sam *PrimarySession) Dial(network, addr string) (net.Conn, error)
```

#### func (*PrimarySession) DialTCP

```go
func (sam *PrimarySession) DialTCP(network string, laddr, raddr net.Addr) (net.Conn, error)
```
DialTCP implements x/dialer

#### func (*PrimarySession) DialTCPI2P

```go
func (sam *PrimarySession) DialTCPI2P(network string, laddr, raddr string) (net.Conn, error)
```

#### func (*PrimarySession) DialUDP

```go
func (sam *PrimarySession) DialUDP(network string, laddr, raddr net.Addr) (net.PacketConn, error)
```
DialUDP implements x/dialer

#### func (*PrimarySession) DialUDPI2P

```go
func (sam *PrimarySession) DialUDPI2P(network, laddr, raddr string) (*DatagramSession, error)
```

#### func (*PrimarySession) From

```go
func (ss *PrimarySession) From() string
```

#### func (*PrimarySession) ID

```go
func (ss *PrimarySession) ID() string
```
Returns the local tunnel name of the I2P tunnel used for the stream session

#### func (*PrimarySession) Keys

```go
func (ss *PrimarySession) Keys() i2pkeys.I2PKeys
```
Returns the keys associated with the stream session

#### func (*PrimarySession) LocalAddr

```go
func (ss *PrimarySession) LocalAddr() net.Addr
```

#### func (*PrimarySession) Lookup

```go
func (s *PrimarySession) Lookup(name string) (a net.Addr, err error)
```

#### func (*PrimarySession) NewDatagramSubSession

```go
func (s *PrimarySession) NewDatagramSubSession(id string, udpPort int) (*DatagramSession, error)
```
Creates a new datagram session. udpPort is the UDP port SAM is listening on, and
if you set it to zero, it will use SAMs standard UDP port.

#### func (*PrimarySession) NewRawSubSession

```go
func (s *PrimarySession) NewRawSubSession(id string, udpPort int) (*RawSession, error)
```
Creates a new raw session. udpPort is the UDP port SAM is listening on, and if
you set it to zero, it will use SAMs standard UDP port.

#### func (*PrimarySession) NewStreamSubSession

```go
func (sam *PrimarySession) NewStreamSubSession(id string) (*StreamSession, error)
```
Creates a new StreamSession with the I2CP- and streaminglib options as
specified. See the I2P documentation for a full list of options.

#### func (*PrimarySession) NewUniqueStreamSubSession

```go
func (sam *PrimarySession) NewUniqueStreamSubSession(id string) (*StreamSession, error)
```
Creates a new StreamSession with the I2CP- and streaminglib options as
specified. See the I2P documentation for a full list of options.

#### func (*PrimarySession) Resolve

```go
func (sam *PrimarySession) Resolve(network, addr string) (net.Addr, error)
```

#### func (*PrimarySession) ResolveTCPAddr

```go
func (sam *PrimarySession) ResolveTCPAddr(network, dest string) (net.Addr, error)
```

#### func (*PrimarySession) ResolveUDPAddr

```go
func (sam *PrimarySession) ResolveUDPAddr(network, dest string) (net.Addr, error)
```

#### func (*PrimarySession) SignatureType

```go
func (ss *PrimarySession) SignatureType() string
```

#### func (*PrimarySession) To

```go
func (ss *PrimarySession) To() string
```

#### type RawSession

```go
type RawSession struct {
}
```

The RawSession provides no authentication of senders, and there is no sender
address attached to datagrams, so all communication is anonymous. The messages
send are however still endpoint-to-endpoint encrypted. You need to figure out a
way to identify and authenticate clients yourself, iff that is needed. Raw
datagrams may be at most 32 kB in size. There is no overhead of authentication,
which is the reason to use this..

#### func (*RawSession) Close

```go
func (s *RawSession) Close() error
```
Closes the RawSession.

#### func (*RawSession) LocalAddr

```go
func (s *RawSession) LocalAddr() i2pkeys.I2PAddr
```
Returns the local I2P destination of the RawSession.

#### func (*RawSession) Read

```go
func (s *RawSession) Read(b []byte) (n int, err error)
```
Reads one raw datagram sent to the destination of the DatagramSession. Returns
the number of bytes read. Who sent the raw message can not be determined at this
layer - you need to do it (in a secure way!).

#### func (*RawSession) SetDeadline

```go
func (s *RawSession) SetDeadline(t time.Time) error
```

#### func (*RawSession) SetReadDeadline

```go
func (s *RawSession) SetReadDeadline(t time.Time) error
```

#### func (*RawSession) SetWriteDeadline

```go
func (s *RawSession) SetWriteDeadline(t time.Time) error
```

#### func (*RawSession) WriteTo

```go
func (s *RawSession) WriteTo(b []byte, addr i2pkeys.I2PAddr) (n int, err error)
```
Sends one raw datagram to the destination specified. At the time of writing,
maximum size is 32 kilobyte, but this may change in the future.

#### type SAM

```go
type SAM struct {
	Config SAMEmit
}
```

Used for controlling I2Ps SAMv3.

#### func  NewSAM

```go
func NewSAM(address string) (*SAM, error)
```
Creates a new controller for the I2P routers SAM bridge.

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

#### func (*SAM) NewDatagramSession

```go
func (s *SAM) NewDatagramSession(id string, keys i2pkeys.I2PKeys, options []string, udpPort int) (*DatagramSession, error)
```
Creates a new datagram session. udpPort is the UDP port SAM is listening on, and
if you set it to zero, it will use SAMs standard UDP port.

#### func (*SAM) NewKeys

```go
func (sam *SAM) NewKeys(sigType ...string) (i2pkeys.I2PKeys, error)
```
Creates the I2P-equivalent of an IP address, that is unique and only the one who
has the private keys can send messages from. The public keys are the I2P
desination (the address) that anyone can send messages to.

#### func (*SAM) NewPrimarySession

```go
func (sam *SAM) NewPrimarySession(id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error)
```
Creates a new PrimarySession with the I2CP- and streaminglib options as
specified. See the I2P documentation for a full list of options.

#### func (*SAM) NewPrimarySessionWithSignature

```go
func (sam *SAM) NewPrimarySessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error)
```
Creates a new PrimarySession with the I2CP- and PRIMARYinglib options as
specified. See the I2P documentation for a full list of options.

#### func (*SAM) NewRawSession

```go
func (s *SAM) NewRawSession(id string, keys i2pkeys.I2PKeys, options []string, udpPort int) (*RawSession, error)
```
Creates a new raw session. udpPort is the UDP port SAM is listening on, and if
you set it to zero, it will use SAMs standard UDP port.

#### func (*SAM) NewStreamSession

```go
func (sam *SAM) NewStreamSession(id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error)
```
Creates a new StreamSession with the I2CP- and streaminglib options as
specified. See the I2P documentation for a full list of options.

#### func (*SAM) NewStreamSessionWithSignature

```go
func (sam *SAM) NewStreamSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error)
```
Creates a new StreamSession with the I2CP- and streaminglib options as
specified. See the I2P documentation for a full list of options.

#### func (*SAM) NewStreamSessionWithSignatureAndPorts

```go
func (sam *SAM) NewStreamSessionWithSignatureAndPorts(id, from, to string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error)
```
Creates a new StreamSession with the I2CP- and streaminglib options as
specified. See the I2P documentation for a full list of options.

#### func (*SAM) ReadKeys

```go
func (sam *SAM) ReadKeys(r io.Reader) (err error)
```
read public/private keys from an io.Reader

#### type SAMConn

```go
type SAMConn struct {
}
```

Implements net.Conn

#### func (*SAMConn) Close

```go
func (sc *SAMConn) Close() error
```
Implements net.Conn

#### func (*SAMConn) LocalAddr

```go
func (sc *SAMConn) LocalAddr() net.Addr
```

#### func (*SAMConn) Read

```go
func (sc *SAMConn) Read(buf []byte) (int, error)
```
Implements net.Conn

#### func (*SAMConn) RemoteAddr

```go
func (sc *SAMConn) RemoteAddr() net.Addr
```

#### func (*SAMConn) SetDeadline

```go
func (sc *SAMConn) SetDeadline(t time.Time) error
```
Implements net.Conn

#### func (*SAMConn) SetReadDeadline

```go
func (sc *SAMConn) SetReadDeadline(t time.Time) error
```
Implements net.Conn

#### func (*SAMConn) SetWriteDeadline

```go
func (sc *SAMConn) SetWriteDeadline(t time.Time) error
```
Implements net.Conn

#### func (*SAMConn) Write

```go
func (sc *SAMConn) Write(buf []byte) (int, error)
```
Implements net.Conn

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

#### type StreamListener

```go
type StreamListener struct {
}
```


#### func (*StreamListener) Accept

```go
func (l *StreamListener) Accept() (net.Conn, error)
```
implements net.Listener

#### func (*StreamListener) AcceptI2P

```go
func (l *StreamListener) AcceptI2P() (*SAMConn, error)
```
accept a new inbound connection

#### func (*StreamListener) Addr

```go
func (l *StreamListener) Addr() net.Addr
```
get our address implements net.Listener

#### func (*StreamListener) Close

```go
func (l *StreamListener) Close() error
```
implements net.Listener

#### func (*StreamListener) From

```go
func (l *StreamListener) From() string
```

#### func (*StreamListener) To

```go
func (l *StreamListener) To() string
```

#### type StreamSession

```go
type StreamSession struct {
	Timeout  time.Duration
	Deadline time.Time
}
```

Represents a streaming session.

#### func (*StreamSession) Addr

```go
func (s *StreamSession) Addr() i2pkeys.I2PAddr
```
Returns the I2P destination (the address) of the stream session

#### func (*StreamSession) Close

```go
func (s *StreamSession) Close() error
```

#### func (*StreamSession) Dial

```go
func (s *StreamSession) Dial(n, addr string) (c net.Conn, err error)
```
implement net.Dialer

#### func (*StreamSession) DialContext

```go
func (s *StreamSession) DialContext(ctx context.Context, n, addr string) (net.Conn, error)
```
context-aware dialer, eventually...

#### func (*StreamSession) DialContextI2P

```go
func (s *StreamSession) DialContextI2P(ctx context.Context, n, addr string) (*SAMConn, error)
```
context-aware dialer, eventually...

#### func (*StreamSession) DialI2P

```go
func (s *StreamSession) DialI2P(addr i2pkeys.I2PAddr) (*SAMConn, error)
```
Dials to an I2P destination and returns a SAMConn, which implements a net.Conn.

#### func (*StreamSession) From

```go
func (s *StreamSession) From() string
```

#### func (*StreamSession) ID

```go
func (s *StreamSession) ID() string
```
Returns the local tunnel name of the I2P tunnel used for the stream session

#### func (*StreamSession) Keys

```go
func (s *StreamSession) Keys() i2pkeys.I2PKeys
```
Returns the keys associated with the stream session

#### func (*StreamSession) Listen

```go
func (s *StreamSession) Listen() (*StreamListener, error)
```
create a new stream listener to accept inbound connections

#### func (*StreamSession) LocalAddr

```go
func (s *StreamSession) LocalAddr() net.Addr
```

#### func (*StreamSession) Lookup

```go
func (s *StreamSession) Lookup(name string) (i2pkeys.I2PAddr, error)
```
lookup name, convenience function

#### func (*StreamSession) Read

```go
func (s *StreamSession) Read(buf []byte) (int, error)
```
Read reads data from the stream.

#### func (*StreamSession) SetDeadline

```go
func (s *StreamSession) SetDeadline(t time.Time) error
```

#### func (*StreamSession) SetReadDeadline

```go
func (s *StreamSession) SetReadDeadline(t time.Time) error
```

#### func (*StreamSession) SetWriteDeadline

```go
func (s *StreamSession) SetWriteDeadline(t time.Time) error
```

#### func (*StreamSession) SignatureType

```go
func (s *StreamSession) SignatureType() string
```

#### func (*StreamSession) To

```go
func (s *StreamSession) To() string
```

#### func (*StreamSession) Write

```go
func (s *StreamSession) Write(data []byte) (int, error)
```
Write sends data over the stream.
