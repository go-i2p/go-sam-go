# sam3
--
    import "github.com/go-i2p/go-sam-go"

Package sam3 provides a compatibility layer for the go-i2p/sam3 library using
go-sam-go as the backend

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
Constants from original sam3

```go
var (
	Options_Humongous = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=3", "outbound.backupQuantity=3",
		"inbound.quantity=6", "outbound.quantity=6",
	}

	Options_Large = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=1", "outbound.backupQuantity=1",
		"inbound.quantity=4", "outbound.quantity=4",
	}

	Options_Wide = []string{
		"inbound.length=1", "outbound.length=1",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=2", "outbound.backupQuantity=2",
		"inbound.quantity=3", "outbound.quantity=3",
	}

	Options_Medium = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=2", "outbound.quantity=2",
	}

	Options_Default = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=0", "outbound.lengthVariance=0",
		"inbound.backupQuantity=1", "outbound.backupQuantity=1",
		"inbound.quantity=1", "outbound.quantity=1",
	}

	Options_Small = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=1", "outbound.quantity=1",
	}

	Options_Warning_ZeroHop = []string{
		"inbound.length=0", "outbound.length=0",
		"inbound.lengthVariance=0", "outbound.lengthVariance=0",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=2", "outbound.quantity=2",
	}
)
```
Predefined option sets (keeping your existing definitions)

```go
var (
	PrimarySessionSwitch string = PrimarySessionString()
	SAM_HOST                    = getEnv("sam_host", "127.0.0.1")
	SAM_PORT                    = getEnv("sam_port", "7656")
)
```
Global variables from original sam3

#### func  ConvertOptionsToSlice

```go
func ConvertOptionsToSlice(opts Options) []string
```
Additional utility functions that may be needed for compatibility

#### func  ExtractDest

```go
func ExtractDest(input string) string
```
ExtractDest extracts destination from input

#### func  ExtractPairInt

```go
func ExtractPairInt(input, value string) int
```
ExtractPairInt extracts integer value from key=value pair

#### func  ExtractPairString

```go
func ExtractPairString(input, value string) string
```
ExtractPairString extracts string value from key=value pair

#### func  GenerateOptionString

```go
func GenerateOptionString(opts []string) string
```
GenerateOptionString generates option string from slice

#### func  GetSAM3Logger

```go
func GetSAM3Logger() *logrus.Logger
```
GetSAM3Logger returns the initialized logger

#### func  IgnorePortError

```go
func IgnorePortError(err error) error
```
IgnorePortError ignores port-related errors

#### func  InitializeSAM3Logger

```go
func InitializeSAM3Logger()
```
InitializeSAM3Logger initializes the logger

#### func  PrimarySessionString

```go
func PrimarySessionString() string
```
PrimarySessionString returns primary session string

#### func  RandString

```go
func RandString() string
```
RandString generates a random string

#### func  SAMDefaultAddr

```go
func SAMDefaultAddr(fallforward string) string
```
SAMDefaultAddr returns default SAM address

#### func  SetAccessListType

```go
func SetAccessListType(s string) func(*I2PConfig) error
```

#### func  SetCloseIdleTime

```go
func SetCloseIdleTime(s string) func(*I2PConfig) error
```

#### func  SetInAllowZeroHop

```go
func SetInAllowZeroHop(s string) func(*I2PConfig) error
```
Configuration option setters for all the missing Set* functions

#### func  SetInBackupQuantity

```go
func SetInBackupQuantity(s string) func(*I2PConfig) error
```

#### func  SetInLength

```go
func SetInLength(s string) func(*I2PConfig) error
```

#### func  SetInQuantity

```go
func SetInQuantity(s string) func(*I2PConfig) error
```

#### func  SetInVariance

```go
func SetInVariance(s string) func(*I2PConfig) error
```

#### func  SetOutAllowZeroHop

```go
func SetOutAllowZeroHop(s string) func(*I2PConfig) error
```

#### func  SetOutBackupQuantity

```go
func SetOutBackupQuantity(s string) func(*I2PConfig) error
```

#### func  SetOutLength

```go
func SetOutLength(s string) func(*I2PConfig) error
```

#### func  SetOutQuantity

```go
func SetOutQuantity(s string) func(*I2PConfig) error
```

#### func  SetOutVariance

```go
func SetOutVariance(s string) func(*I2PConfig) error
```

#### func  SetReduceIdleTime

```go
func SetReduceIdleTime(s string) func(*I2PConfig) error
```

#### func  SetUseCompression

```go
func SetUseCompression(s string) func(*I2PConfig) error
```

#### type DatagramSession

```go
type DatagramSession struct {
}
```

DatagramSession implements net.PacketConn for I2P datagrams

#### func (*DatagramSession) Accept

```go
func (s *DatagramSession) Accept() (net.Conn, error)
```
Accept accepts connections (not applicable for datagrams)

#### func (*DatagramSession) Addr

```go
func (s *DatagramSession) Addr() net.Addr
```
Addr returns the session address

#### func (*DatagramSession) B32

```go
func (s *DatagramSession) B32() string
```
B32 returns the base32 address

#### func (*DatagramSession) Close

```go
func (s *DatagramSession) Close() error
```
Close closes the datagram session

#### func (*DatagramSession) Dial

```go
func (s *DatagramSession) Dial(net string, addr string) (*DatagramSession, error)
```
Dial dials a connection (returns self for datagrams)

#### func (*DatagramSession) DialI2PRemote

```go
func (s *DatagramSession) DialI2PRemote(net string, addr net.Addr) (*DatagramSession, error)
```
DialI2PRemote dials to I2P remote

#### func (*DatagramSession) DialRemote

```go
func (s *DatagramSession) DialRemote(net, addr string) (net.PacketConn, error)
```
DialRemote dials to remote address

#### func (*DatagramSession) LocalAddr

```go
func (s *DatagramSession) LocalAddr() net.Addr
```
LocalAddr returns the local address

#### func (*DatagramSession) LocalI2PAddr

```go
func (s *DatagramSession) LocalI2PAddr() i2pkeys.I2PAddr
```
LocalI2PAddr returns the I2P destination

#### func (*DatagramSession) Lookup

```go
func (s *DatagramSession) Lookup(name string) (a net.Addr, err error)
```
Lookup performs name lookup

#### func (*DatagramSession) Read

```go
func (s *DatagramSession) Read(b []byte) (n int, err error)
```
Read reads from the session

#### func (*DatagramSession) ReadFrom

```go
func (s *DatagramSession) ReadFrom(b []byte) (n int, addr net.Addr, err error)
```
ReadFrom reads a datagram from the session

#### func (*DatagramSession) RemoteAddr

```go
func (s *DatagramSession) RemoteAddr() net.Addr
```
RemoteAddr returns the remote address

#### func (*DatagramSession) SetDeadline

```go
func (s *DatagramSession) SetDeadline(t time.Time) error
```
SetDeadline sets read and write deadlines

#### func (*DatagramSession) SetReadDeadline

```go
func (s *DatagramSession) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets read deadline

#### func (*DatagramSession) SetWriteBuffer

```go
func (s *DatagramSession) SetWriteBuffer(bytes int) error
```
SetWriteBuffer sets write buffer size

#### func (*DatagramSession) SetWriteDeadline

```go
func (s *DatagramSession) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets write deadline

#### func (*DatagramSession) Write

```go
func (s *DatagramSession) Write(b []byte) (int, error)
```
Write writes to the session

#### func (*DatagramSession) WriteTo

```go
func (s *DatagramSession) WriteTo(b []byte, addr net.Addr) (n int, err error)
```
WriteTo writes a datagram to the specified address

#### type I2PConfig

```go
type I2PConfig struct {
	*common.I2PConfig
}
```

I2PConfig manages I2P configuration options

#### func  NewConfig

```go
func NewConfig(opts ...func(*I2PConfig) error) (*I2PConfig, error)
```
NewConfig creates a new I2PConfig

#### func (*I2PConfig) DestinationKey

```go
func (f *I2PConfig) DestinationKey() string
```

#### func (*I2PConfig) ID

```go
func (f *I2PConfig) ID() string
```

#### func (*I2PConfig) MaxSAM

```go
func (f *I2PConfig) MaxSAM() string
```

#### func (*I2PConfig) MinSAM

```go
func (f *I2PConfig) MinSAM() string
```

#### func (*I2PConfig) Print

```go
func (f *I2PConfig) Print() []string
```

#### func (*I2PConfig) Reduce

```go
func (f *I2PConfig) Reduce() string
```

#### func (*I2PConfig) Reliability

```go
func (f *I2PConfig) Reliability() string
```

#### func (*I2PConfig) SAMAddress

```go
func (f *I2PConfig) SAMAddress() string
```

#### func (*I2PConfig) Sam

```go
func (f *I2PConfig) Sam() string
```

#### func (*I2PConfig) SessionStyle

```go
func (f *I2PConfig) SessionStyle() string
```

#### func (*I2PConfig) SetSAMAddress

```go
func (f *I2PConfig) SetSAMAddress(addr string)
```
All the configuration method forwards

#### func (*I2PConfig) SignatureType

```go
func (f *I2PConfig) SignatureType() string
```

#### func (*I2PConfig) ToPort

```go
func (f *I2PConfig) ToPort() string
```

#### type Option

```go
type Option func(*SAMEmit) error
```

Option is a functional option for SAMEmit

#### type Options

```go
type Options map[string]string
```

Options represents a map of configuration options

#### func  ConvertSliceToOptions

```go
func ConvertSliceToOptions(slice []string) Options
```

#### func (Options) AsList

```go
func (opts Options) AsList() (ls []string)
```
AsList returns options as a list of strings

#### type PrimarySession

```go
type PrimarySession struct {
	Timeout  time.Duration
	Deadline time.Time
	Config   SAMEmit
}
```

PrimarySession represents a primary session

#### func (*PrimarySession) Addr

```go
func (ss *PrimarySession) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P address

#### func (*PrimarySession) Close

```go
func (ss *PrimarySession) Close() error
```
Close closes the session

#### func (*PrimarySession) Dial

```go
func (sam *PrimarySession) Dial(network, addr string) (net.Conn, error)
```
Dial implements net.Dialer

#### func (*PrimarySession) DialTCP

```go
func (sam *PrimarySession) DialTCP(network string, laddr, raddr net.Addr) (net.Conn, error)
```
DialTCP implements x/dialer

#### func (*PrimarySession) DialTCPI2P

```go
func (sam *PrimarySession) DialTCPI2P(network string, laddr, raddr string) (net.Conn, error)
```
DialTCPI2P dials TCP over I2P

#### func (*PrimarySession) DialUDP

```go
func (sam *PrimarySession) DialUDP(network string, laddr, raddr net.Addr) (net.PacketConn, error)
```
DialUDP implements x/dialer

#### func (*PrimarySession) DialUDPI2P

```go
func (sam *PrimarySession) DialUDPI2P(network, laddr, raddr string) (*DatagramSession, error)
```
DialUDPI2P dials UDP over I2P

#### func (*PrimarySession) From

```go
func (ss *PrimarySession) From() string
```
From returns from port

#### func (*PrimarySession) ID

```go
func (ss *PrimarySession) ID() string
```
ID returns the session ID

#### func (*PrimarySession) Keys

```go
func (ss *PrimarySession) Keys() i2pkeys.I2PKeys
```
Keys returns the session keys

#### func (*PrimarySession) LocalAddr

```go
func (ss *PrimarySession) LocalAddr() net.Addr
```
LocalAddr returns local address

#### func (*PrimarySession) Lookup

```go
func (s *PrimarySession) Lookup(name string) (a net.Addr, err error)
```
Lookup performs name lookup

#### func (*PrimarySession) NewDatagramSubSession

```go
func (s *PrimarySession) NewDatagramSubSession(id string, udpPort int) (*DatagramSession, error)
```
NewDatagramSubSession creates a new datagram sub-session

#### func (*PrimarySession) NewRawSubSession

```go
func (s *PrimarySession) NewRawSubSession(id string, udpPort int) (*RawSession, error)
```
NewRawSubSession creates a new raw sub-session

#### func (*PrimarySession) NewStreamSubSession

```go
func (sam *PrimarySession) NewStreamSubSession(id string) (*StreamSession, error)
```
NewStreamSubSession creates a new stream sub-session

#### func (*PrimarySession) NewStreamSubSessionWithPorts

```go
func (sam *PrimarySession) NewStreamSubSessionWithPorts(id, from, to string) (*StreamSession, error)
```
NewStreamSubSessionWithPorts creates a new stream sub-session with ports

#### func (*PrimarySession) NewUniqueStreamSubSession

```go
func (sam *PrimarySession) NewUniqueStreamSubSession(id string) (*StreamSession, error)
```
NewUniqueStreamSubSession creates a unique stream sub-session

#### func (*PrimarySession) Resolve

```go
func (sam *PrimarySession) Resolve(network, addr string) (net.Addr, error)
```
Resolve resolves network address

#### func (*PrimarySession) ResolveTCPAddr

```go
func (sam *PrimarySession) ResolveTCPAddr(network, dest string) (net.Addr, error)
```
ResolveTCPAddr resolves TCP address

#### func (*PrimarySession) ResolveUDPAddr

```go
func (sam *PrimarySession) ResolveUDPAddr(network, dest string) (net.Addr, error)
```
ResolveUDPAddr resolves UDP address

#### func (*PrimarySession) SignatureType

```go
func (ss *PrimarySession) SignatureType() string
```
SignatureType returns signature type

#### func (*PrimarySession) To

```go
func (ss *PrimarySession) To() string
```
To returns to port

#### type RawSession

```go
type RawSession struct {
}
```

RawSession provides raw datagram messaging

#### func (*RawSession) Close

```go
func (s *RawSession) Close() error
```
Close closes the raw session

#### func (*RawSession) LocalAddr

```go
func (s *RawSession) LocalAddr() i2pkeys.I2PAddr
```
LocalAddr returns the local I2P destination

#### func (*RawSession) Read

```go
func (s *RawSession) Read(b []byte) (n int, err error)
```
Read reads one raw datagram

#### func (*RawSession) SetDeadline

```go
func (s *RawSession) SetDeadline(t time.Time) error
```
SetDeadline sets read and write deadlines

#### func (*RawSession) SetReadDeadline

```go
func (s *RawSession) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets read deadline

#### func (*RawSession) SetWriteDeadline

```go
func (s *RawSession) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets write deadline

#### func (*RawSession) WriteTo

```go
func (s *RawSession) WriteTo(b []byte, addr i2pkeys.I2PAddr) (n int, err error)
```
WriteTo sends one raw datagram to the destination

#### type SAM

```go
type SAM struct {
	Config SAMEmit
}
```

SAM represents the main controller for I2P router's SAM bridge

#### func  NewSAM

```go
func NewSAM(address string) (*SAM, error)
```
NewSAM creates a new controller for the I2P routers SAM bridge

#### func (*SAM) Close

```go
func (sam *SAM) Close() error
```
Close closes this sam session

#### func (*SAM) EnsureKeyfile

```go
func (sam *SAM) EnsureKeyfile(fname string) (keys i2pkeys.I2PKeys, err error)
```
EnsureKeyfile ensures keyfile exists

#### func (*SAM) Keys

```go
func (sam *SAM) Keys() (k *i2pkeys.I2PKeys)
```
Keys returns the keys associated with this SAM instance

#### func (*SAM) Lookup

```go
func (sam *SAM) Lookup(name string) (i2pkeys.I2PAddr, error)
```
Lookup performs a name lookup

#### func (*SAM) NewDatagramSession

```go
func (s *SAM) NewDatagramSession(id string, keys i2pkeys.I2PKeys, options []string, udpPort int) (*DatagramSession, error)
```
NewDatagramSession creates a new datagram session

#### func (*SAM) NewKeys

```go
func (sam *SAM) NewKeys(sigType ...string) (i2pkeys.I2PKeys, error)
```
NewKeys creates the I2P-equivalent of an IP address

#### func (*SAM) NewPrimarySession

```go
func (sam *SAM) NewPrimarySession(id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error)
```
NewPrimarySession creates a new PrimarySession

#### func (*SAM) NewPrimarySessionWithSignature

```go
func (sam *SAM) NewPrimarySessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error)
```
NewPrimarySessionWithSignature creates a new PrimarySession with signature

#### func (*SAM) NewRawSession

```go
func (s *SAM) NewRawSession(id string, keys i2pkeys.I2PKeys, options []string, udpPort int) (*RawSession, error)
```
NewRawSession creates a new raw session

#### func (*SAM) NewStreamSession

```go
func (sam *SAM) NewStreamSession(id string, keys i2pkeys.I2PKeys, options []string) (*StreamSession, error)
```
NewStreamSession creates a new StreamSession

#### func (*SAM) NewStreamSessionWithSignature

```go
func (sam *SAM) NewStreamSessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error)
```
NewStreamSessionWithSignature creates a new StreamSession with custom signature

#### func (*SAM) NewStreamSessionWithSignatureAndPorts

```go
func (sam *SAM) NewStreamSessionWithSignatureAndPorts(id, from, to string, keys i2pkeys.I2PKeys, options []string, sigType string) (*StreamSession, error)
```
NewStreamSessionWithSignatureAndPorts creates a new StreamSession with signature
and ports

#### func (*SAM) ReadKeys

```go
func (sam *SAM) ReadKeys(r io.Reader) (err error)
```
ReadKeys reads public/private keys from an io.Reader

#### type SAMConn

```go
type SAMConn struct {
}
```

SAMConn implements net.Conn for I2P connections

#### func (*SAMConn) Close

```go
func (sc *SAMConn) Close() error
```
Close closes the connection

#### func (*SAMConn) LocalAddr

```go
func (sc *SAMConn) LocalAddr() net.Addr
```
LocalAddr returns the local address

#### func (*SAMConn) Read

```go
func (sc *SAMConn) Read(buf []byte) (int, error)
```
Read reads data from the connection

#### func (*SAMConn) RemoteAddr

```go
func (sc *SAMConn) RemoteAddr() net.Addr
```
RemoteAddr returns the remote address

#### func (*SAMConn) SetDeadline

```go
func (sc *SAMConn) SetDeadline(t time.Time) error
```
SetDeadline sets read and write deadlines

#### func (*SAMConn) SetReadDeadline

```go
func (sc *SAMConn) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets read deadline

#### func (*SAMConn) SetWriteDeadline

```go
func (sc *SAMConn) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets write deadline

#### func (*SAMConn) Write

```go
func (sc *SAMConn) Write(buf []byte) (int, error)
```
Write writes data to the connection

#### type SAMEmit

```go
type SAMEmit struct {
	I2PConfig
}
```

SAMEmit handles SAM protocol message generation

#### func  NewEmit

```go
func NewEmit(opts ...func(*SAMEmit) error) (*SAMEmit, error)
```
NewEmit creates a new SAMEmit

#### func (*SAMEmit) Accept

```go
func (e *SAMEmit) Accept() string
```
Accept generates accept message

#### func (*SAMEmit) AcceptBytes

```go
func (e *SAMEmit) AcceptBytes() []byte
```
AcceptBytes generates accept message as bytes

#### func (*SAMEmit) Connect

```go
func (e *SAMEmit) Connect(dest string) string
```
Connect generates connect message

#### func (*SAMEmit) ConnectBytes

```go
func (e *SAMEmit) ConnectBytes(dest string) []byte
```
ConnectBytes generates connect message as bytes

#### func (*SAMEmit) Create

```go
func (e *SAMEmit) Create() string
```
Create generates session create message

#### func (*SAMEmit) CreateBytes

```go
func (e *SAMEmit) CreateBytes() []byte
```
CreateBytes generates session create message as bytes

#### func (*SAMEmit) GenerateDestination

```go
func (e *SAMEmit) GenerateDestination() string
```
GenerateDestination generates destination message

#### func (*SAMEmit) GenerateDestinationBytes

```go
func (e *SAMEmit) GenerateDestinationBytes() []byte
```
GenerateDestinationBytes generates destination message as bytes

#### func (*SAMEmit) Hello

```go
func (e *SAMEmit) Hello() string
```
Hello generates hello message

#### func (*SAMEmit) HelloBytes

```go
func (e *SAMEmit) HelloBytes() []byte
```
HelloBytes generates hello message as bytes

#### func (*SAMEmit) Lookup

```go
func (e *SAMEmit) Lookup(name string) string
```
Lookup generates lookup message

#### func (*SAMEmit) LookupBytes

```go
func (e *SAMEmit) LookupBytes(name string) []byte
```
LookupBytes generates lookup message as bytes

#### func (*SAMEmit) SamOptionsString

```go
func (e *SAMEmit) SamOptionsString() string
```
SamOptionsString returns SAM options as string

#### type SAMResolver

```go
type SAMResolver struct {
	*SAM
}
```

SAMResolver provides name resolution functionality

#### func  NewFullSAMResolver

```go
func NewFullSAMResolver(address string) (*SAMResolver, error)
```
NewFullSAMResolver creates a new full SAMResolver

#### func  NewSAMResolver

```go
func NewSAMResolver(parent *SAM) (*SAMResolver, error)
```
NewSAMResolver creates a new SAMResolver from existing SAM

#### func (*SAMResolver) Resolve

```go
func (sam *SAMResolver) Resolve(name string) (i2pkeys.I2PAddr, error)
```
Resolve performs a lookup

#### type StreamListener

```go
type StreamListener struct {
}
```

StreamListener implements net.Listener for I2P streams

#### func (*StreamListener) Accept

```go
func (l *StreamListener) Accept() (net.Conn, error)
```
Accept accepts new inbound connections

#### func (*StreamListener) AcceptI2P

```go
func (l *StreamListener) AcceptI2P() (*SAMConn, error)
```
AcceptI2P accepts a new inbound I2P connection

#### func (*StreamListener) Addr

```go
func (l *StreamListener) Addr() net.Addr
```
Addr returns the listener's address

#### func (*StreamListener) Close

```go
func (l *StreamListener) Close() error
```
Close closes the listener

#### func (*StreamListener) From

```go
func (l *StreamListener) From() string
```
From returns the from port

#### func (*StreamListener) To

```go
func (l *StreamListener) To() string
```
To returns the to port

#### type StreamSession

```go
type StreamSession struct {
	Timeout  time.Duration
	Deadline time.Time
}
```

StreamSession represents a streaming session

#### func (*StreamSession) Addr

```go
func (s *StreamSession) Addr() i2pkeys.I2PAddr
```
Addr returns the I2P destination address

#### func (*StreamSession) Close

```go
func (s *StreamSession) Close() error
```
Close closes the session

#### func (*StreamSession) Dial

```go
func (s *StreamSession) Dial(n, addr string) (c net.Conn, err error)
```
Dial establishes a connection to an address

#### func (*StreamSession) DialContext

```go
func (s *StreamSession) DialContext(ctx context.Context, n, addr string) (net.Conn, error)
```
DialContext establishes a connection with context

#### func (*StreamSession) DialContextI2P

```go
func (s *StreamSession) DialContextI2P(ctx context.Context, n, addr string) (*SAMConn, error)
```
DialContextI2P establishes an I2P connection with context

#### func (*StreamSession) DialI2P

```go
func (s *StreamSession) DialI2P(addr i2pkeys.I2PAddr) (*SAMConn, error)
```
DialI2P dials to an I2P destination

#### func (*StreamSession) From

```go
func (s *StreamSession) From() string
```
From returns the from port

#### func (*StreamSession) ID

```go
func (s *StreamSession) ID() string
```
ID returns the local tunnel name

#### func (*StreamSession) Keys

```go
func (s *StreamSession) Keys() i2pkeys.I2PKeys
```
Keys returns the keys associated with the session

#### func (*StreamSession) Listen

```go
func (s *StreamSession) Listen() (*StreamListener, error)
```
Listen creates a new stream listener

#### func (*StreamSession) LocalAddr

```go
func (s *StreamSession) LocalAddr() net.Addr
```
LocalAddr returns the local address

#### func (*StreamSession) Lookup

```go
func (s *StreamSession) Lookup(name string) (i2pkeys.I2PAddr, error)
```
Lookup performs name lookup

#### func (*StreamSession) Read

```go
func (s *StreamSession) Read(buf []byte) (int, error)
```
Read reads data from the stream

#### func (*StreamSession) RemoteAddr

```go
func (s *StreamSession) RemoteAddr() net.Addr
```
RemoteAddr returns the remote address

#### func (*StreamSession) SetDeadline

```go
func (s *StreamSession) SetDeadline(t time.Time) error
```
SetDeadline sets read and write deadlines

#### func (*StreamSession) SetReadDeadline

```go
func (s *StreamSession) SetReadDeadline(t time.Time) error
```
SetReadDeadline sets read deadline

#### func (*StreamSession) SetWriteDeadline

```go
func (s *StreamSession) SetWriteDeadline(t time.Time) error
```
SetWriteDeadline sets write deadline

#### func (*StreamSession) SignatureType

```go
func (s *StreamSession) SignatureType() string
```
SignatureType returns the signature type

#### func (*StreamSession) To

```go
func (s *StreamSession) To() string
```
To returns the to port

#### func (*StreamSession) Write

```go
func (s *StreamSession) Write(data []byte) (int, error)
```
Write sends data over the stream
