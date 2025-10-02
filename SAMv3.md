# SAM V3

*This page was last updated in 2025-04 and is accurate for router version API 0.9.66.*

SAM is a simple client protocol for interacting with I2P. SAM is the recommended protocol for non-Java applications to connect to the I2P network, and is supported by multiple router implementations. Java applications should use the streaming or I2CP APIs directly.

SAM version 3 was introduced in I2P release 0.7.3 (May 2009) and is a stable and supported interface. 3.1 is also stable and supports the signature type option, which is strongly recommended. More recent 3.x versions support advanced features. Note that i2pd does not currently support most 3.2 and 3.3 features.

**Alternatives:** SOCKS, Streaming, I2CP, BOB (deprecated)  
**Deprecated versions:** SAM V1, SAM V2

## Known SAM Libraries

⚠️ **Warning:** Some of these may be very old or unsupported. None are tested, reviewed, or maintained by the I2P project unless noted below. Do your own research.

| Library Name | Language | Version | STREAM | DGRAM | RAW | Site |
|-------------|----------|---------|---------|--------|-----|------|
| i2psam | C++, C wrapper | 3.1 | yes | yes | no | github.com/i2p/i2psam |
| gosam | Go | 3.2 | yes | no | no | github.com/eyedeekay/goSam |
| sam3 | Go | 3.3 | yes | yes | yes | github.com/eyedeekay/sam3 |
| onramp | Go | 3.3 | yes | yes | yes | github.com/eyedeekay/onramp |
| txi2p | Python | 3.1 | yes | no | no | github.com/str4d/txi2p |
| i2p.socket | Python | 3.2 | yes | yes | yes | github.com/majestrate/i2p.socket |
| i2plib | Python | 3.1 | yes | no | no | github.com/l-n-s/i2plib |
| i2plib-fork | Python | 3.1 | yes | no | no | codeberg.org/weko/i2plib-fork |
| Py2p | Python | 3.3 | yes | yes | yes | i2pgit.org/robin/Py2p |
| i2p-rs | Rust | 3.1 | yes | yes | yes | github.com/i2p/i2p-rs |
| libsam3 | C | 3.1 | yes | yes | yes | github.com/i2p/libsam3 *(Maintained by the I2P project)* |
| mooni2p | Lua | 3.1 | yes | yes | yes | notabug.org/villain/mooni2p |
| haskell-network-anonymous-i2p | Haskell | 3.1 | yes | yes | yes | github.com/solatis/haskell-network-anonymous-i2p |
| i2p-sam | Javascript | 3.1 | yes | no | yes | codeberg.org/diva.exchange/i2p-sam |
| node-i2p | Javascript | 3.0 | yes | unk | unk | github.com/redhog/node-i2p |
| Jsam | Java | 3.1 | yes | no | no | github.com/eyedeekay/Jsam |
| I2PSharp | .Net | 3.3 | yes | no | no | github.com/MohA39/I2PSharp |
| i2pdotnet | .Net | 3.0 | yes | unk | unk | github.com/SamuelFisher/i2pdotnet |
| i2p.rb | Ruby | 3.0 | yes | no | no | github.com/dryruby/i2p.rb |
| solitude | Rust | 3.1 | WIP | WIP | WIP | github.com/syvita/solitude |
| Samty | C++ | 3.1 | yes | no | no | notabug.org/acetone/samty |
| bitcoin | C++ | 3.1 | yes | no | no | source (not a library, but good reference code) |

## Quick Start

To implement a basic TCP-only, peer-to-peer application, the client must support the following commands:

1. **HELLO VERSION MIN=3.1 MAX=3.1**  
   Needed for all of the remaining ones

2. **DEST GENERATE SIGNATURE_TYPE=7**  
   To generate our private key and destination

3. **NAMING LOOKUP NAME=...**  
   To convert .i2p addresses to destinations

4. **SESSION CREATE STYLE=STREAM ID=... DESTINATION=... i2cp.leaseSetEncType=4,0**  
   Needed for STREAM CONNECT and STREAM ACCEPT

5. **STREAM CONNECT ID=... DESTINATION=...**  
   To make outgoing connections

6. **STREAM ACCEPT ID=...**  
   To accept incoming connections

## General Guidance for Developers

### Application Design

SAM sessions (or inside I2P, tunnel pools or sets of tunnels) are designed to be long-lived. Most applications will only need one session, created at startup and closed on exit. I2P is different from Tor, where circuits may be rapidly created and discarded. Think carefully and consult with I2P developers before designing your application to use more than one or two simultaneous sessions, or to rapidly create and discard them. Most threat models will not require a unique session for every connection.

Also, please ensure your application settings (and guidance to users about router settings, or router defaults if you bundle a router) will result in your users contributing more resources to the network than they consume. I2P is a peer-to-peer network, and the network cannot survive if a popular application drives the network into permanent congestion.

### Compatibility and Testing

The Java I2P and i2pd router implementations are independent and have minor differences in behavior, feature support, and defaults. Please test your application with the latest version of both routers.

i2pd SAM is enabled by default; Java I2P SAM is not. Provide instructions to your users on how to enable SAM in Java I2P (via /configclients in the router console), and/or provide a good error message to the user if the initial connect fails, e.g. "ensure that I2P is running and the SAM interface is enabled".

The Java I2P and i2pd routers have different defaults for tunnel quantities. The Java default is 2 and the i2pd default is 5. For most low- to medium-bandwidth and low- to medium-connection counts, 2 or 3 is sufficient. Please specify the tunnel quantity in the SESSION CREATE message to get consistent performance with the Java I2P and i2pd routers. See below.

For more guidance to developers on ensuring your application uses only the resources it needs, please see our guide to bundling I2P with your application.

### Signature and Encryption Types

I2P supports multiple signature and encryption types. For backward compatibility, SAM defaults to old and inefficient types, so all clients should specify newer types.

The signature type is specified in the DEST GENERATE and SESSION CREATE (for transient) commands. All clients should set `SIGNATURE_TYPE=7` (Ed25519).

The encryption type is specified in the SESSION CREATE command. Multiple encryption types are allowed. Clients should set either `i2cp.leaseSetEncType=4` (for ECIES-X25519 only) or `i2cp.leaseSetEncType=4,0` (for ECIES-X25519 and ElGamal, if compatibility is required).

## Version 3 Changes

### Version 3.0 Changes

Version 3.0 was introduced in I2P release 0.7.3. SAM v2 provided a way to manage several sockets on the same I2P destination in parallel, i.e. the client does not have to wait for data being successfully sent on one socket before sending data on another socket. But all data transited through the same client<-->SAM socket, which was quite complicated to manage for the client.

SAM v3 manages sockets in a different way: each I2P socket matches a unique client<-->SAM socket, which is much more simple to handle. This is similar to BOB. SAM v3 also offers a UDP port for sending datagrams through I2P, and can forward back I2P datagrams to the client's datagram server.

### Version 3.1 Changes

Version 3.1 was introduced in Java I2P release 0.9.14 (July 2014). SAM 3.1 is the recommended minimum SAM implementation because of its support for better signature types than SAM 3.0. i2pd also supports most 3.1 features.

- DEST GENERATE and SESSION CREATE now support a SIGNATURE_TYPE parameter
- The MIN and MAX parameters in HELLO VERSION are now optional
- The MIN and MAX parameters in HELLO VERSION now support single-digit versions such as "3"
- RAW SEND is now supported on the bridge socket

### Version 3.2 Changes

Version 3.2 was introduced in Java I2P release 0.9.24 (January 2016). Note that i2pd does not currently support most 3.2 features.

#### I2CP Port and Protocol Support

- SESSION CREATE options FROM_PORT and TO_PORT
- SESSION CREATE STYLE=RAW option PROTOCOL
- STREAM CONNECT, DATAGRAM SEND, and RAW SEND options FROM_PORT and TO_PORT
- RAW SEND option PROTOCOL
- DATAGRAM RECEIVED, RAW RECEIVED, and forwarded or received streams and repliable datagrams, includes FROM_PORT and TO_PORT
- RAW session option HEADER=true will cause the forwarded raw datagrams to be prepended with a line containing PROTOCOL=nnn FROM_PORT=nnnn TO_PORT=nnnn
- The first line of datagrams sent through port 7655 may now start with any 3.x version
- The first line of datagrams sent through port 7655 may contain any of the options FROM_PORT, TO_PORT, PROTOCOL
- RAW RECEIVED includes PROTOCOL=nnn

#### SSL and Authentication

- USER/PASSWORD in the HELLO parameters for authorization
- Optional authorization configuration with the AUTH command
- Optional SSL/TLS support on the control socket
- STREAM FORWARD option SSL=true

#### Multithreading

- Concurrent pending STREAM ACCEPTs are allowed on the same session ID

#### Command Line Parsing and Keepalive

- Optional commands QUIT, STOP and EXIT to close the session and socket
- Command parsing will properly handle UTF-8
- Command parsing reliably handles whitespace inside quotes
- A backslash '\' may escape quotes on the command line
- Recommended that the server map commands to upper case, for ease in testing via telnet
- Empty option values such as PROTOCOL or PROTOCOL= may be allowed, implementation dependent
- PING/PONG for keepalive
- Servers may implement timeouts for the HELLO or subsequent commands, implementation dependent

### Version 3.3 Changes

Version 3.3 was introduced in Java I2P release 0.9.25 (March 2016). Note that i2pd does not currently support most 3.3 features.

- The same session may be used for streams, datagrams, and raw simultaneously. Incoming packets and streams will be routed based on I2P protocol and to-port. See the PRIMARY section below.
- DATAGRAM SEND and RAW SEND now support options SEND_TAGS, TAG_THRESHOLD, EXPIRES, and SEND_LEASESET. See the datagram sending section below.

## Version 3 Protocol

### Simple Anonymous Messaging (SAM) Version 3.3 Specification Overview

The client application talks to the SAM bridge, which deals with all of the I2P functionality (using the streaming library for virtual streams, or I2CP directly for datagrams).

By default, the client<-->SAM bridge communication is unencrypted and unauthenticated. The SAM bridge may support SSL/TLS connections; configuration and implementation details are outside the scope of this specification. As of SAM 3.2, optional authentication user/password parameters are supported in the initial handshake and may be required by the bridge.

I2P communications can take several distinct forms:

- Virtual streams
- Repliable and authenticated datagrams (messages with a FROM field)
- Anonymous datagrams (raw anonymous messages)
- Datagram2 (a new repliable and authenticated format)
- Datagram3 (a new repliable but unauthenticated format)

I2P communications are supported by I2P sessions, and each I2P session is bound to an address (called destination). An I2P session is associated with one of the three types above, and cannot carry communications of another type, unless using PRIMARY sessions.

### Encoding and Escaping

All of these SAM messages are sent on a single line, terminated by the newline character (\n). Prior to SAM 3.2, only 7-bit ASCII was supported. As of SAM 3.2, the encoding must be UTF-8. Any UTF8-encoded keys or values should work.

The formatting shown in this specification below is merely for readability, and while the first two words in each message must stay in their specific order, the ordering of the key=value pairs can change (e.g. "ONE TWO A=B C=D" or "ONE TWO C=D A=B" are both perfectly valid constructions). In addition, the protocol is case-sensitive. In the following, message examples are preceded by "-> " for messages sent by the client to the SAM bridge, and by "<- " for messages sent by the SAM bridge to the client.

The basic command or response line takes one of the following forms:

```
COMMAND SUBCOMMAND [key=value] [key=value] ...
COMMAND                                           # As of SAM 3.2
PING[ arbitrary text]                             # As of SAM 3.2
PONG[ arbitrary text]                             # As of SAM 3.2
```

COMMAND without a SUBCOMMAND is supported for some new commands in SAM 3.2 only.

Key=value pairs must be separated by a single space. (As of SAM 3.2, multiple spaces are allowed) Values must be enclosed in double quotes if they contain spaces, e.g. `key="long value text"`. (Prior to SAM 3.2, this did not work reliably in some implementations)

Prior to SAM 3.2, there was no escaping mechanism. As of SAM 3.2, double quotes may be escaped with a backslash '\' and a backslash may be represented as two backslashes '\\'.

#### Empty values

As of SAM 3.2, empty option values such as KEY, KEY=, or KEY="" may be allowed, implementation dependent.

#### Case Sensitivity

The protocol, as specified, is case-sensitive. It is recommended but not required that the server map commands to upper case, for ease in testing via telnet. This would allow, for example, "hello version" to work. This is implementation-dependent. Do not map keys or values to upper case, as this would corrupt I2CP options.

### SAM Connection Handshake

No SAM communication can occur until after the client and bridge have agreed on a protocol version, which is done by the client sending a HELLO and the bridge sending a HELLO REPLY:

```
->  HELLO VERSION
          [MIN=$min]            # Optional as of SAM 3.1, required for 3.0 and earlier
          [MAX=$max]            # Optional as of SAM 3.1, required for 3.0 and earlier
          [USER="xxx"]          # As of SAM 3.2, required if authentication is enabled
          [PASSWORD="yyy"]      # As of SAM 3.2, required if authentication is enabled
```

and

```
<-  HELLO REPLY RESULT=OK VERSION=3.1
```

As of version 3.1 (I2P 0.9.14), the MIN and MAX parameters are optional. SAM will always return the highest version possible given the MIN and MAX constraints, or the current server version if no constraints are given. If the SAM bridge cannot find a suitable version, it replies with:

```
<- HELLO REPLY RESULT=NOVERSION
```

If some error occurred, such as a bad request format, it replies with:

```
<- HELLO REPLY RESULT=I2P_ERROR MESSAGE="$message"
```

#### SSL

The server's control socket may optionally offer SSL/TLS support, as configured on the server and client. Implementations may offer other transport layers as well; this is outside the scope of the protocol definition.

#### Authorization

For authorization, client adds `USER="xxx" PASSWORD="yyy"` to the HELLO parameters. Double quotes for user and password are recommended but not required. A double quote inside a user or password must be escaped with a backslash. On failure the server will reply with an I2P_ERROR and a message. It is recommended that SSL be enabled on any SAM servers where authorization is required.

#### Timeouts

Servers may implement timeouts for the HELLO or subsequent commands, implementation dependent. Clients should promptly send the HELLO and the next command after connecting.

If a timeout occurs before the HELLO is received, the bridge replies with:

```
<- HELLO REPLY RESULT=I2P_ERROR MESSAGE="$message"
```

and then disconnects.

If a timeout occurs after the HELLO is received but before the next command, the bridge replies with:

```
<- SESSION STATUS RESULT=I2P_ERROR MESSAGE="$message"
```

and then disconnects.

### I2CP Ports and Protocol

As of SAM 3.2, the I2CP ports and protocol may be specified by the SAM client sender to be passed through to I2CP, and the SAM bridge will pass the received I2CP port and protocol information to the SAM client.

For FROM_PORT and TO_PORT, the valid range is 0-65535, and the default is 0.

For PROTOCOL, which may be specified only for RAW, the valid range is 0-255, and the default is 18.

For SESSION commands, the specified ports and protocol are the defaults for that session. For individual streams or datagrams, the specified ports and protocol override the session defaults. For received streams or datagrams, the indicated ports and protocol are as received from I2CP.

#### Important Differences from Standard IP

I2CP ports are for I2P sockets and datagrams. They are unrelated to your local sockets connecting to SAM.

- Port 0 is valid and has special meaning
- Ports 1-1023 are not special or privileged
- Servers listen on port 0 by default, which means "all ports"
- Clients send to port 0 by default, which means "any port"
- Clients send from port 0 by default, which means "unspecified"
- Servers may have a service listening on port 0 and other services listening on higher ports. If so, the port 0 service is the default, and will be connected to if the incoming socket or datagram port does not match another service
- Most I2P destinations only have one service running on them, so you may use the defaults, and ignore I2CP port configuration
- SAM 3.2 or 3.3 is required to specify I2CP ports
- If you don't need I2CP ports, you don't need SAM 3.2 or 3.3; 3.1 is sufficient
- Protocol 0 is valid and means "any protocol". This is not recommended, and probably will not work
- I2P sockets are tracked by an internal connection ID. Therefore, there is no requirement that the 5-tuple of dest:port:dest:port:protocol be unique. For example, there may be multiple sockets with the same ports between two destinations. Clients do not need to pick a "free port" for an outbound connection

If you are designing a SAM 3.3 application with multiple subsessions, think carefully about how to use ports and protocols effectively. See the I2CP specification for more information.

### SAM Sessions

A SAM session is created by a client opening a socket to the SAM bridge, operating a handshake, and sending a SESSION CREATE message, and the session terminates when the socket is disconnected.

Each registered I2P Destination is uniquely associated with a session ID (or nickname). Session IDs, including subsession IDs for PRIMARY sessions, must be globally unique on the SAM server. To prevent possible ID collisions with other clients, best practice is for the client to generate IDs randomly.

Each session is uniquely associated with:
- the socket from which the client creates the session
- its ID (or nickname)

#### Session Creation Request

The session creation message can only use one of these forms (messages received through other forms are answered with an error message):

```
->  SESSION CREATE
          STYLE={STREAM,DATAGRAM,RAW,DATAGRAM2,DATAGRAM3}   # See below for DATAGRAM2/3
          ID=$nickname
          DESTINATION={$privkey,TRANSIENT}
          [SIGNATURE_TYPE=value]               # SAM 3.1 or higher only, for DESTINATION=TRANSIENT only, default DSA_SHA1
          [PORT=$port]                         # Required for DATAGRAM* RAW, invalid for STREAM
          [HOST=$host]                         # Optional for DATAGRAM* and RAW, invalid for STREAM
          [FROM_PORT=nnn]                      # SAM 3.2 or higher only, default 0
          [TO_PORT=nnn]                        # SAM 3.2 or higher only, default 0
          [PROTOCOL=nnn]                       # SAM 3.2 or higher only, for STYLE=RAW only, default 18
                                               # 6, 17, 19, 20 not allowed
          [HEADER={true,false}]                # SAM 3.2 or higher only, for STYLE=RAW only, default false
          [sam.udp.host=hostname]              # Datagram bind host, Java I2P only, DATAGRAM*/RAW only, default 127.0.0.1
          [sam.udp.port=nnn]                   # Datagram bind port, Java I2P only, DATAGRAM*/RAW only, default 7655
          [option=value]*                      # I2CP and streaming options
```

DESTINATION specifies what destination should be used for sending and receiving messages/streams. The $privkey is the base 64 of the concatenation of the Destination followed by the Private Key followed by the Signing Private Key, optionally followed by the Offline Signature, which is 663 or more bytes in binary and 884 or more bytes in base 64, depending on signature type. The binary format is specified in Private Key File. See additional notes about the Private Key in the Destination Key Generation section below.

If the signing private key is all zeros, the Offline Signature section follows. Offline signatures are only supported for STREAM and RAW sessions. Offline signatures may not be created with DESTINATION=TRANSIENT. The format of the offline signature section is:

- Expires timestamp (4 bytes, big endian, seconds since epoch, rolls over in 2106)
- Sig type of transient Signing Public Key (2 bytes, big endian)
- Transient Signing Public key (length as specified by transient sig type)
- Signature of above three fields by offline key (length as specified by destination sig type)
- Transient Signing Private key (length as specified by transient sig type)

If the destination is specified as TRANSIENT, the SAM bridge creates a new destination. As of version 3.1 (I2P 0.9.14), if the destination is TRANSIENT, an optional parameter SIGNATURE_TYPE is supported. The SIGNATURE_TYPE value may be any name (e.g. ECDSA_SHA256_P256, case insensitive) or number (e.g. 1) supported by Key Certificates. The default is DSA_SHA1, which is NOT what you want. For most applications, please specify `SIGNATURE_TYPE=7`.

$nickname is the choice of the client. No whitespace is allowed.

Additional options given are passed to the I2P session configuration if not interpreted by the SAM bridge (e.g. `outbound.length=0`).

The Java I2P and i2pd routers have different defaults for tunnel quantities. The Java default is 2 and the i2pd default is 5. For most low- to medium-bandwidth and low- to medium-connection counts, 2 or 3 is sufficient. Please specify the tunnel quantities in the SESSION CREATE message to get consistent performance with the Java I2P and i2pd routers, using the options e.g. `inbound.quantity=3 outbound.quantity=3`. These and other options are documented in the links below.

The SAM bridge itself should already be configured with what router it should communicate over I2P through (though if need be there may be a way to provide an override, e.g. `i2cp.tcp.host=localhost` and `i2cp.tcp.port=7654`).

#### Session Creation Response

After receiving the session create message, the SAM bridge will reply with a session status message, as follows:

If the creation was successful:

```
<-  SESSION STATUS RESULT=OK DESTINATION=$privkey
```

The $privkey is the base 64 of the concatenation of the Destination followed by the Private Key followed by the Signing Private Key, optionally followed by the Offline Signature, which is 663 or more bytes in binary and 884 or more bytes in base 64, depending on signature type. The binary format is specified in Private Key File.

If the SESSION CREATE contained a signing private key of all zeros and a Offline Signature section, the SESSION STATUS reply will include the same data in the same format. See the SESSION CREATE section above for details.

If the nickname is already associated with a session:

```
<-  SESSION STATUS RESULT=DUPLICATED_ID
```

If the destination is already in use:

```
<-  SESSION STATUS RESULT=DUPLICATED_DEST
```

If the destination is not a valid private destination key:

```
<-  SESSION STATUS RESULT=INVALID_KEY
```

If some other error has occurred:

```
<-  SESSION STATUS RESULT=I2P_ERROR MESSAGE="$message"
```

If it's not OK, the MESSAGE should contain human-readable information as to why the session could not be created.

Note that the router builds tunnels before responding with SESSION STATUS. This could take several seconds, or, at router startup or during severe network congestion, a minute or more. If unsuccessful, the router will not respond with a failure message for several minutes. Do not set a short timeout waiting for the response. Do not abandon the session while tunnel build is in progress and retry.

SAM sessions live and die with the socket they are associated with. When the socket is closed, the session dies, and all communications using the session die at the same time. And the other way around, when the session dies for any reason, the SAM bridge closes the socket.

## SAM Virtual Streams

Virtual streams are guaranteed to be sent reliably and in order, with failure and success notification as soon as it is available.

Streams are bidirectional communication sockets between two I2P destinations, but their opening has to be requested by one of them. Hereafter, CONNECT commands are used by the SAM client for such a request. FORWARD / ACCEPT commands are used by the SAM client when he wants to listen to requests coming from other I2P destinations.

### SAM Virtual Streams: CONNECT

A client asks for a connection by:

- opening a new socket with the SAM bridge
- passing the same HELLO handshake as above
- sending the STREAM CONNECT command

#### Connect Request

```
-> STREAM CONNECT
         ID=$nickname
         DESTINATION=$destination
         [SILENT={true,false}]                # default false
         [FROM_PORT=nnn]                      # SAM 3.2 or higher only, default 0
         [TO_PORT=nnn]                        # SAM 3.2 or higher only, default 0
```

This establishes a new virtual connection from the local session whose ID is $nickname to the specified peer.

The target is $destination, which is the base 64 of the Destination, which is 516 or more base 64 characters (387 or more bytes in binary), depending on signature type. NOTE: Since about 2014 (SAM v3.1), Java I2P has also supported hostnames and b32 addresses for the $destination, but this was previously undocumented. Hostnames and b32 addresses are now officially supported by Java I2P as of release 0.9.48. The i2pd router supports hostnames and b32 addresses as of release 2.38.0 (0.9.50). For both routers, "b32" support includes support extended "b33" addresses for blinded destinations.

#### Connect Response

If SILENT=true is passed, the SAM bridge won't issue any other message on the socket. If the connection fails, the socket will be closed. If the connection succeeds, all remaining data passing through the current socket is forwarded from and to the connected I2P destination peer.

If SILENT=false, which is the default value, the SAM bridge sends a last message to its client before forwarding or shutting down the socket:

```
<-  STREAM STATUS
         RESULT=$result
         [MESSAGE=...]
```

The RESULT value may be one of:
- OK
- CANT_REACH_PEER
- I2P_ERROR
- INVALID_KEY
- INVALID_ID
- TIMEOUT

If the RESULT is OK, all remaining data passing through the current socket is forwarded from and to the connected I2P destination peer. If the connection was not possible (timeout, etc), RESULT will contain the appropriate error value (accompanied by an optional human-readable MESSAGE), and the SAM bridge closes the socket.

The router stream connect timeout internally is approximately one minute, implementation-dependent. Do not set a shorter timeout waiting for the response.

### SAM Virtual Streams: ACCEPT

A client waits for an incoming connection request by:

- opening a new socket with the SAM bridge
- passing the same HELLO handshake as above
- sending the STREAM ACCEPT command

#### Accept Request

```
-> STREAM ACCEPT
         ID=$nickname
         [SILENT={true,false}]                # default false
```

This makes the session ${nickname} listen for one incoming connection request from the I2P network. ACCEPT is not allowed while there is an active FORWARD on the session.

As of SAM 3.2, multiple concurrent pending STREAM ACCEPTs are allowed on the same session ID (even with the same port). Prior to 3.2, concurrent accepts would fail with ALREADY_ACCEPTING. Note: Java I2P also supports concurrent ACCEPTs on SAM 3.1, as of release 0.9.24 (2016-01). i2pd also supports concurrent ACCEPTs on SAM 3.1, as of release 2.50.0 (2023-12).

#### Accept Response

If SILENT=true is passed, the SAM bridge won't issue any other message on the socket. If the accept fails, the socket will be closed. If the accept succeeds, all remaining data passing through the current socket is forwarded from and to the connected I2P destination peer. For reliability, and to receive the destination for incoming connections, SILENT=false is recommended.

If SILENT=false, which is the default value, the SAM bridge answers with:

```
<-  STREAM STATUS
         RESULT=$result
         [MESSAGE=...]
```

The RESULT value may be one of:
- OK
- I2P_ERROR
- INVALID_ID

If the result is not OK, the socket is closed immediately by the SAM bridge. If the result is OK, the SAM bridge starts waiting for an incoming connection request from another I2P peer. When a request arrives, the SAM bridge accepts it and:

If SILENT=true was passed, the SAM bridge won't issue any other message on the client socket. All remaining data passing through the current socket is forwarded from and to the connected I2P destination peer.

If SILENT=false was passed, which is the default value, the SAM bridge sends the client a ASCII line containing the base64 public destination key of the requesting peer, and additional information for SAM 3.2 only:

```
$destination FROM_PORT=nnn TO_PORT=nnn
```

After this '\n' terminated line, all remaining data passing through the current socket is forwarded from and to the connected I2P destination peer, until one of the peer closes the socket.

#### Errors After OK

In rare cases, the SAM bridge may encounter an error after sending RESULT=OK, but before a connection comes in and sending the $destination line to the client. These errors may include router shutdown, router restart, and session close. In these cases, when SILENT=false, the SAM bridge may, but is not required to (implementation-dependent), send the line:

```
<-  STREAM STATUS
         RESULT=I2P_ERROR
         [MESSAGE=...]
```

before immediately closing the socket. This line is not, of course, decodable as a valid Base 64 destination.

### SAM Virtual Streams: FORWARD

A client can use a regular socket server and wait for connection requests coming from I2P. For that, the client must:

- open a new socket with the SAM bridge
- pass the same HELLO handshake as above
- send the forward command

#### Forward Request

```
-> STREAM FORWARD
         ID=$nickname
         PORT=$port
         [HOST=$host]
         [SILENT={true,false}]                # default false
         [SSL={true,false}]                   # SAM 3.2 or higher only, default false
```

This makes the session ${nickname} listen for incoming connection requests from the I2P network. FORWARD is not allowed while there is a pending ACCEPT on the session.

#### Forward Response

SILENT defaults to false. Whether SILENT is true or false, the SAM bridge always answers with a STREAM STATUS message. Note that this is a different behavior from STREAM ACCEPT and STREAM CONNECT when SILENT=true. The STREAM STATUS message is:

```
<-  STREAM STATUS
         RESULT=$result
         [MESSAGE=...]
```

The RESULT value may be one of:
- OK
- I2P_ERROR
- INVALID_ID

$host is the hostname or IP address of the socket server to which SAM will forward connection requests. If not given, SAM takes the IP of the socket that issued the forward command.

$port is the port number of the socket server to which SAM will forward connection requests. It is mandatory.

When a connection request arrives from I2P, the SAM bridge opens a socket connection to $host:$port. If it is accepted in less than 3 seconds, SAM will accept the connection from I2P, and then:

If SILENT=true was passed, all data passing through the obtained current socket is forwarded from and to the connected I2P destination peer.

If SILENT=false was passed, which is the default value, the SAM bridge sends on the obtained socket an ASCII line containing the base64 public destination key of the requesting peer, and additional information for SAM 3.2 only:

```
$destination FROM_PORT=nnn TO_PORT=nnn
```

After this '\n' terminated line, all remaining data passing through the socket is forwarded from and to the connected I2P destination peer, until one of the sides closes the socket.

As of SAM 3.2, if SSL=true is specified, the forwarding socket is over SSL/TLS.

The I2P router will stop listening to incoming connection requests as soon as the "forwarding" socket is closed.

## SAM Datagrams

SAMv3 provides mechanisms to send and receive datagrams over local datagram sockets. Some SAMv3 implementations also support the older v1/v2 way of sending/receiving datagrams over the SAM bridge socket. Both are documented below.

I2P supports four types of datagrams:

- Repliable and authenticated datagrams are prefixed with the destination of the sender, and contain the signature of the sender, so the receiver may verify that the sender's destination was not spoofed, and may reply to the datagram. The new Datagram2 format is also repliable and authenticated.
- The new Datagram3 format is repliable but not authenticated. The sender information is unverified.
- Raw datagrams do not contain the destination of the sender or a signature.

Default I2CP ports are defined for both repliable and raw datagrams. The I2CP port may be changed for raw datagrams.

A common protocol design pattern is for repliable datagrams to be sent to servers, with some identifier included, and the server to respond with a raw datagram that includes that indentifier, so the response may be correlated with the request. This design pattern eliminates the substantial overhead of repliable datagrams in replies. All choices of I2CP protocols and ports are application-specific, and designers should take these issues into consideration.

See also the important notes on datagram MTU in the section below.

### Sending Repliable or Raw Datagrams

While I2P doesn't inherently contain a FROM address, for ease of use an additional layer is provided as repliable datagrams - unordered and unreliable messages of up to 31744 bytes that include a FROM address (leaving up to 1KB for header material). This FROM address is authenticated internally by SAM (making use of the destination's signing key to verify the source) and includes replay prevention.

Minimum size is 1. For best delivery reliability, recommended maximum size is approximately 11 KB. Reliability is inversely proportional to message size, perhaps even exponentially.

After establishing a SAM session with STYLE=DATAGRAM or STYLE=RAW, the client can send repliable or raw datagrams through SAM's UDP port (7655 by default).

The first line of a datagram sent through this port must be in the following format. This is all on one line (space separated), shown on multiple lines for clarity:

```
3.0                                  # As of SAM 3.2, any "3.x" is allowed. Prior to that, "3.0" is required.
$nickname
$destination
[FROM_PORT=nnn]                      # SAM 3.2 or higher only, default from session options
[TO_PORT=nnn]                        # SAM 3.2 or higher only, default from session options
[PROTOCOL=nnn]                       # SAM 3.2 or higher only, only for RAW sessions, default from session options
[SEND_TAGS=nnn]                      # SAM 3.3 or higher only, number of session tags to send
                                     # Overrides crypto.tagsToSend I2CP session option
                                     # Default is router-dependent (40 for Java router)
[TAG_THRESHOLD=nnn]                  # SAM 3.3 or higher only, low session tag threshold
                                     # Overrides crypto.lowTagThreshold I2CP session option
                                     # Default is router-dependent (30 for Java router)
[EXPIRES=nnn]                        # SAM 3.3 or higher only, expiration from now in seconds
                                     # Overrides clientMessageTimeout I2CP session option (which is in ms)
                                     # Default is router-dependent (60 for Java router)
[SEND_LEASESET={true,false}]         # SAM 3.3 or higher only, whether to send our leaseset
                                     # Overrides shouldBundleReplyInfo I2CP session option
                                     # Default is true
```

- 3.0 is the version of SAM. As of SAM 3.2, any 3.x is allowed.
- $nickname is the id of the DATAGRAM session that will be used
- The target is $destination, which is the base 64 of the Destination, which is 516 or more base 64 characters (387 or more bytes in binary), depending on signature type. NOTE: Since about 2014 (SAM v3.1), Java I2P has also supported hostnames and b32 addresses for the $destination, but this was previously undocumented. Hostnames and b32 addresses are now officially supported by Java I2P as of release 0.9.48. The i2pd router does not currently support hostnames and b32 addresses; support may be added in a future release.
- All options are per-datagram settings that override the defaults specified in the SESSION CREATE.
- Version 3.3 options SEND_TAGS, TAG_THRESHOLD, EXPIRES, and SEND_LEASESET will be passed to I2CP if supported. See the I2CP specification for details. Support by the SAM server is optional, it will ignore these options if unsupported.
- this line is '\n' terminated.

The first line will be discarded by SAM before sending the remaining data of the message to the specified destination.

For an alternate method of sending repliable and raw datagrams, see DATAGRAM SEND and RAW SEND.

### SAM Repliable Datagrams: Receiving a Datagram

Received datagrams are written by SAM on the socket from which the datagram session was opened, if a forwarding PORT is not specified in the SESSION CREATE command. This is the v1/v2-compatible way of receiving datagrams.

When a datagram arrives, the bridge delivers it to the client via the message:

```
<-  DATAGRAM RECEIVED
           DESTINATION=$destination           # See notes below for Datagram3 format
           SIZE=$numBytes
           FROM_PORT=nnn                      # SAM 3.2 or higher only
           TO_PORT=nnn                        # SAM 3.2 or higher only
           \n
       [$numBytes of data]
```

The source is $destination, which is the base 64 of the Destination, which is 516 or more base 64 characters (387 or more bytes in binary), depending on signature type.

The SAM bridge never exposes to the client the authentication headers or other fields, merely the data that the sender provided. This continues until the session is closed (by the client dropping the connection).

### Forwarding Raw or Repliable Datagrams

When creating a datagram session, the client can ask SAM to forward incoming messages to a specified ip:port. It does so by issuing the CREATE command with PORT and HOST options:

```
-> SESSION CREATE
          STYLE={DATAGRAM,RAW,DATAGRAM2,DATAGRAM3}   # See below for DATAGRAM2/3
          ID=$nickname
          DESTINATION={$privkey,TRANSIENT}
          PORT=$port
          [HOST=$host]
          [FROM_PORT=nnn]                      # SAM 3.2 or higher only, default 0
          [TO_PORT=nnn]                        # SAM 3.2 or higher only, default 0
          [PROTOCOL=nnn]                       # SAM 3.2 or higher only, for STYLE=RAW only, default 18
                                               # 6, 17, 19, 20 not allowed
          [sam.udp.host=hostname]              # Datagram bind host, Java I2P only, default 127.0.0.1
          [sam.udp.port=nnn]                   # Datagram bind port, Java I2P only, default 7655
          [option=value]*                      # I2CP options
```

The $privkey is the base 64 of the concatenation of the Destination followed by the Private Key followed by the Signing Private Key, optionally followed by the Offline Signature, which is 884 or more base 64 characters (663 or more bytes in binary), depending on signature type. The binary format is specified in Private Key File.

Offline signatures are supported for RAW, DATAGRAM2, and DATAGRAM3 datagrams, but not for DATAGRAM. See the SESSION CREATE section above and the DATAGRAM2/3 section below for details.

$host is the hostname or IP address of the datagram server to which SAM will forward datagrams. If not given, SAM takes the IP of the socket that issued the forward command.

$port is the port number of the datagram server to which SAM will forward datagrams. If $port is not set, datagrams will NOT be forwarded, they will be received on the control socket, in the v1/v2-compatible way.

Additional options given are passed to the I2P session configuration if not interpreted by the SAM bridge (e.g. `outbound.length=0`). These options are documented below.

Forwarded repliable datagrams are always prefixed with the base64 destination, except for Datagram3, see below. When a repliable datagram arrives, the bridge sends to the specified host:port a UDP packet containing the following data:

```
$destination                       # See notes below for Datagram3 format
FROM_PORT=nnn                      # SAM 3.2 or higher only
TO_PORT=nnn                        # SAM 3.2 or higher only
\n
$datagram_payload
```

Forwarded raw datagrams are forwarded as-is to the specified host:port without a prefix. The UDP packet contains the following data:

```
$datagram_payload
```

As of SAM 3.2, when HEADER=true is specified in SESSION CREATE, the forwarded raw datagram will be prepended with a header line as follows:

```
FROM_PORT=nnn
TO_PORT=nnn
PROTOCOL=nnn
\n
$datagram_payload
```

The $destination is the base 64 of the Destination, which is 516 or more base 64 characters (387 or more bytes in binary), depending on signature type.

### SAM Anonymous (Raw) Datagrams

Squeezing the most out of I2P's bandwidth, SAM allows clients to send and receive anonymous datagrams, leaving authentication and reply information up to the client themselves. These datagrams are unreliable and unordered, and may be up to 32768 bytes.

Minimum size is 1. For best delivery reliability, recommended maximum size is approximately 11 KB.

After establishing a SAM session with STYLE=RAW, the client can send anonymous datagrams through the SAM bridge exactly the same way as sending repliable datagrams.

Both ways of receiving datagrams are also available for anonymous datagrams.

Received datagrams are written by SAM on the socket from which the datagram session was opened, if a forwarding PORT is not specified in the SESSION CREATE command. This is the v1/v2-compatible way of receiving datagrams.

```
<- RAW RECEIVED
          SIZE=$numBytes
          FROM_PORT=nnn                      # SAM 3.2 or higher only
          TO_PORT=nnn                        # SAM 3.2 or higher only
          PROTOCOL=nnn                       # SAM 3.2 or higher only
          \n
      [$numBytes of data]
```

When anonymous datagrams are to be forwarded to some host:port, the bridge sends to the specified host:port a message containing the following data:

```
$datagram_payload
```

As of SAM 3.2, when HEADER=true is specified in SESSION CREATE, the forwarded raw datagram will be prepended with a header line as follows:

```
FROM_PORT=nnn
TO_PORT=nnn
PROTOCOL=nnn
\n
$datagram_payload
```

For an alternate method of sending anonymous datagrams, see RAW SEND.

### Datagram 2/3

Datagram 2/3 are new formats specified in early 2025. No known implementations currently exist. Check implementation documentation for current status. See the specification for more information.

There are no current plans to increase the SAM version to indicate Datagram 2/3 support. This may be problematic as implementations may wish to support Datagram 2/3 but not SAM v3.3 features. Any version change is TBD.

Both Datagram2 and Datagram3 are repliable. Only Datagram2 is authenticated.

Datagram2 is identical to repliable datagrams from a SAM perspective. Both are authenticated. Only the I2CP format and signature are different, but this is not visible to SAM clients. Datagram2 also supports offline signatures, so it may be used by offline-signed destinations.

The intention is for Datagram2 to replace Repliable datagrams for new applications that do not require backward-compatibility. Datagram2 provides replay protection that is not present for Repliable datagrams. If backward-compatibility is required, an application may support both Datagram2 and Repliable may be supported on the same session with SAM 3.3 PRIMARY sessions.

Datagram3 is repliable but not authenticated. The 'from' field in the I2CP format is a hash, not a destination. The $destination as sent from the SAM server to the client will be a 44-byte base64 hash. To convert it to a full destination for reply, base64-decode it to 32 bytes binary, then base32-encode it to 52 characters and append ".b32.i2p" for a NAMING LOOKUP. As usual, clients should maintain their own cache to avoid repeated NAMING LOOKUPs.

Application designers should use extreme caution and consider the security implications of unauthenticated datagrams.

### V3 Datagram MTU Considerations

I2P Datagrams may be larger than the typical internet MTU of 1500. Locally sent datagrams and forwarded repliable datagrams prefixed with the 516+ byte base64 destination are likely to exceed that MTU. However, localhost MTUs on Linux systems are typically much larger, for example 65536. Localhost MTUs will vary by OS. I2P Datagrams will never be larger than 65536. Datagram size is dependent on the application protocol.

If the SAM client is local to the SAM server and the system supports a larger MTU, then the datagrams will not be fragmented locally. However, if the SAM client is remote, then IPv4 datagrams would be fragmented and IPv6 datagrams would fail (IPv6 does not support UDP fragmentation).

Client library and application developers should be aware of these issues and document recommendations to avoid fragmentation and prevent packet loss, especially on remote SAM client-server connections.

### DATAGRAM SEND, RAW SEND (V1/V2 Compatible Datagram Handling)

In SAM V3, the preferred way to send datagrams is via the datagram socket at port 7655 as documented above. However, repliable datagrams may be sent directly via the SAM bridge socket using the DATAGRAM SEND command, as documented in SAM V1 and SAM V2.

As of release 0.9.14 (version 3.1), anonymous datagrams may be sent directly via the SAM bridge socket using the RAW SEND command, as documented in SAM V1 and SAM V2.

As of release 0.9.24 (version 3.2), DATAGRAM SEND and RAW SEND may include the parameters FROM_PORT=nnnn and/or TO_PORT=nnnn to override the default ports. As of release 0.9.24 (version 3.2), RAW SEND may include the parameter PROTOCOL=nnn to override the default protocol.

These commands do not support the ID parameter. The datagrams are sent to the most recently created DATAGRAM- or RAW-style session, as appropriate. Support for the ID parameter may be added in a future release.

DATAGRAM2 and DATAGRAM3 formats are not supported in the V1/V2 compatible way.

## SAM PRIMARY Sessions (V3.3 and higher)

Version 3.3 was introduced in I2P release 0.9.25.

In an earlier version of this specification, PRIMARY sessions were known as MASTER sessions. In both i2pd and I2P+, they are still known only as MASTER sessions.

SAM v3.3 adds support for running streaming, datagrams, and raw subsessions on the same primary session, and for running multiple subsessions of the same style. All subsession traffic uses a single destination, or set of tunnels. Routing of traffic from I2P is based on the port and protocol options for the subsessions.

To create multiplexed subsessions, you must create a primary session and then add subsessions to the primary session. Each subsession must have a unique id and a unique listen protocol and port. Subsessions may also be removed from the primary session.

With a PRIMARY session and a combination of subsessions, a SAM client may support multiple applications, or a single sophisticated application using a variety of protocols, on a single set of tunnels. For example, a bittorrent client could set up a streaming subsession for peer-to-peer connections, together with datagram and raw subsessions for DHT communication.

### Creating a PRIMARY Session

```
->  SESSION CREATE
          STYLE=PRIMARY                        # prior to 0.9.47, use STYLE=MASTER
          ID=$nickname
          DESTINATION={$privkey,TRANSIENT}
          [sam.udp.host=hostname]              # Datagram bind host, Java I2P only, default 127.0.0.1
          [sam.udp.port=nnn]                   # Datagram bind port, Java I2P only, default 7655
          [option=value]*                      # I2CP and streaming options
```

The SAM bridge will respond with success or failure as in the response to a standard SESSION CREATE.

Do not set the PORT, HOST, FROM_PORT, TO_PORT, PROTOCOL, LISTEN_PORT, LISTEN_PROTOCOL, or HEADER options on a primary session. You may not send any data on a PRIMARY session ID or on the control socket. All commands such as STREAM CONNECT, DATAGRAM SEND, etc. must use the subsession ID on a separate socket.

The PRIMARY session connects to the router and builds tunnels. When the SAM bridge responds, tunnels have been built and the session is ready for subsessions to be added. All I2CP options pertaining to tunnel parameters such as length, quantity, and nickname must be provided in the primary's SESSION CREATE.

All utility commands are supported on a primary session.

When the primary session is closed, all subsessions get closed also.

**NOTE:** Prior to release 0.9.47, use STYLE=MASTER. STYLE=PRIMARY is supported as of release 0.9.47. MASTER is still supported for backwards compatibility.

### Creating a Subsession

Using the same control socket on which the PRIMARY session was created:

```
->  SESSION ADD
          STYLE={STREAM,DATAGRAM,RAW,DATAGRAM2,DATAGRAM3}   # See above for DATAGRAM2/3
          ID=$nickname                         # must be unique
          [PORT=$port]                         # Required for DATAGRAM* and RAW, invalid for STREAM
          [HOST=$host]                         # Optional for DATAGRAM* and RAW, invalid for STREAM
          [FROM_PORT=nnn]                      # For outbound traffic, default 0
          [TO_PORT=nnn]                        # For outbound traffic, default 0
          [PROTOCOL=nnn]                       # For outbound traffic for STYLE=RAW only, default 18
                                               # 6, 17, 19, 20 not allowed
          [LISTEN_PORT=nnn]                    # For inbound traffic, default is the FROM_PORT value
                                               # For STYLE=STREAM, only the FROM_PORT value or 0 is allowed
          [LISTEN_PROTOCOL=nnn]                # For inbound traffic for STYLE=RAW only
                                               # Default is the PROTOCOL value; 6 (streaming) is disallowed
          [HEADER={true,false}]                # For STYLE=RAW only, default false
          [sam.udp.host=hostname]              # Datagram bind host, Java I2P only, DATAGRAM*/RAW only, default 127.0.0.1
          [sam.udp.port=nnn]                   # Datagram bind port, Java I2P only, DATAGRAM*/RAW only, default 7655
          [option=value]*                      # I2CP and streaming options
```

The SAM bridge will respond with success or failure as in the response to a standard SESSION CREATE. As the tunnels were already built in the primary SESSION CREATE, the SAM bridge should respond immediately.

Do not set the DESTINATION option on a SESSION ADD. The subsession will use the destination specified in the primary session. All subsessions must be added on the control socket, i.e. the same connection that you created the primary session on.

Multiple subsessions must have options sufficiently unique that incoming data can be routed correctly. In particular, multiple sessions of the same style must have different LISTEN_PORT options (and/or LISTEN_PROTOCOL, for RAW only). A SESSION ADD with listen port and protocol that duplicates an existing subsession will result in an error.

The LISTEN_PORT is the local I2P port, i.e. the receive (TO) port for incoming data. If the LISTEN_PORT is not specified, the FROM_PORT value will be used. If the LISTEN_PORT and FROM_PORT are not specified, incoming routing will be based on STYLE and PROTOCOL alone. For LISTEN_PORT and LISTEN_PROTOCOL, 0 means any value, that is, a wildcard. If both LISTEN_PORT and LISTEN_PROTOCOL are 0, this subsession will be the default for incoming traffic that does not get routed to another subsession. Incoming streaming traffic (protocol 6) will never be routed to a RAW subsession, even if its LISTEN_PROTCOL is 0. A RAW subsession may not set a LISTEN_PROTOCOL of 6. If there is no default or subsession that matching the protocol and port of incoming traffic, that data will be dropped.

Use the subsession ID, not the primary session ID, for sending and receiving data. All commands such as STREAM CONNECT, DATAGRAM SEND, etc. must use the subsession ID.

All utility commands are supported on a primary session or subsession. v1/v2 datagram/raw sending/receiving are not supported on a primary session or on subsessions.

### Stopping a Subsession

Using the same control socket on which the PRIMARY session was created:

```
->  SESSION REMOVE
          ID=$nickname
```

This removes a subsession from the primary session. Do not set any other options on a SESSION REMOVE. Subsessions must be removed on the control socket, i.e. the same connection that you created the primary session on. After a subsession is removed, it is closed and may not be used to send or receive data.

The SAM bridge will respond with success or failure as in the response to a standard SESSION CREATE.

## SAM Utility Commands

Some utility commands require a pre-existing session and some do not. See details below.

### Host Name Lookup

The following message can be used by the client to query the SAM bridge for name resolution:

```
NAMING LOOKUP
        NAME=$name
        [OPTIONS=true]     # Default false, as of router API 0.9.66
```

which is answered by

```
NAMING REPLY
        RESULT=$result
        NAME=$name
        [VALUE=$destination]
        [MESSAGE="$message"]
        [OPTION:optionkey="$optionvalue"]   # As of router API 0.9.66
```

The RESULT value may be one of:
- OK
- INVALID_KEY
- KEY_NOT_FOUND

If NAME=ME, then the reply will contain the destination used by the current session (useful if you're using a TRANSIENT one). If $result is not OK, MESSAGE may convey a descriptive message, such as "bad format", etc. INVALID_KEY implies that something is wrong with $name in the request, possibly invalid characters.

The $destination is the base 64 of the Destination, which is 516 or more base 64 characters (387 or more bytes in binary), depending on signature type.

NAMING LOOKUP does not require that a session has been created first. However, in some implementations, a .b32.i2p lookup which is uncached and requires a network query may fail, as no client tunnels are available for the lookup.

#### Name Lookup Options

NAMING LOOKUP is extended as of router API 0.9.66 to support service lookups. Support may vary by implementation. See proposal 167 for additional information.

NAMING LOOKUP NAME=example.i2p OPTIONS=true requests the options mapping in the reply. NAME may be a full base64 destination when OPTIONS=true.

If the destination lookup was successful and options were present in the leaseset, then in the reply, following the destination, will be one or more options in the form of OPTION:key=value. Each option will have a separate OPTION: prefix. All options from the leaseset will be included, not just service record options. For example, options for parameters defined in the future may be present. Example:

```
NAMING REPLY RESULT=OK NAME=example.i2p VALUE=base64dest OPTION:_smtp._tcp="1 86400 0 0 25 bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb.b32.i2p"
```

Keys containing '=', and keys or values containing a newline, are considered invalid and the key/value pair will be removed from the reply. If there are no options found in the leaseset, or if the leaseset was version 1, then the response will not include any options. If OPTIONS=true was in the lookup, and the leaseset is not found, a new result value LEASESET_NOT_FOUND will be returned.

### Destination Key Generation

Public and private base64 keys can be generated using the following message:

```
->  DEST GENERATE
          [SIGNATURE_TYPE=value]               # SAM 3.1 or higher only, default DSA_SHA1
```

which is answered by

```
DEST REPLY
      PUB=$destination
      PRIV=$privkey
```

As of version 3.1 (I2P 0.9.14), an optional parameter SIGNATURE_TYPE is supported. The SIGNATURE_TYPE value may be any name (e.g. ECDSA_SHA256_P256, case insensitive) or number (e.g. 1) that is supported by Key Certificates. The default is DSA_SHA1, which is NOT what you want. For most applications, please specify `SIGNATURE_TYPE=7`.

The $destination is the base 64 of the Destination, which is 516 or more base 64 characters (387 or more bytes in binary), depending on signature type.

The $privkey is the base 64 of the concatenation of the Destination followed by the Private Key followed by the Signing Private Key, which is 884 or more base 64 characters (663 or more bytes in binary), depending on signature type. The binary format is specified in Private Key File.

**Notes about the 256-byte binary Private Key:** This field has been unused since version 0.6 (2005). SAM implementations may send random data or all zeros in this field; do not be alarmed about a string of AAAA in the base 64. Most applications will simply store the base 64 string and return it as-is in the SESSION CREATE, or decode to binary for storage, then encode again for SESSION CREATE. Applications may, however, decode the base 64, parse the binary following the PrivateKeyFile specification, discard the 256-byte private key portion, and then replace it with 256 bytes of random data or all zeros when re-encoding it for the SESSION CREATE. ALL other fields in the PrivateKeyFile specification must be preserved. This would save 256 bytes of file system storage but is probably not worth the trouble for most applications. See proposal 161 for addtional information and background.

DEST GENERATE does not require that a session has been created first.

DEST GENERATE cannot be used to create a destination with offline signatures.

### PING/PONG (SAM 3.2 or higher)

Either the client or server may send:

```
PING[ arbitrary text]
```

on the control port, with the response:

```
PONG[ arbitrary text from the ping]
```

to be used for control socket keepalive. Either side may close the session and socket if no response is received in a reasonable time, implementation dependent.

If a timeout occurs waiting for a PONG from the client, the bridge may send:

```
<- SESSION STATUS RESULT=I2P_ERROR MESSAGE="$message"
```

and then disconnect.

If a timeout occurs waiting for a PONG from the bridge, the client may simply disconnect.

PING/PONG do not require that a session has been created first.

### QUIT/STOP/EXIT (SAM 3.2 or higher, optional features)

Commands QUIT, STOP and EXIT will close the session and socket. Implementation is optional, for ease in testing via telnet. Whether there is any response before the socket is closed (for example, a SESSION STATUS message) is implementation-specific and outside the scope of this specification.

QUIT/STOP/EXIT do not require that a session has been created first.

### HELP (optional feature)

Servers may implement a HELP command. Implementation is optional, for ease in testing via telnet. Output format and detection of the end of the output is implementation-specific and outside the scope of this specification.

HELP does not require that a session has been created first.

### Authorization Configuration (SAM 3.2 or higher, optional feature)

Authorization configuration using the AUTH command. A SAM server may implement these commands to facilite persistent storage of credentials. Configuration of authentication other than with these commands is implementation-specific and outside the scope of this specification.

- `AUTH ENABLE` enables authorization on subsequent connections
- `AUTH DISABLE` disables authorization on subsequent connections
- `AUTH ADD USER="foo" PASSWORD="bar"` adds a user/password
- `AUTH REMOVE USER="foo"` removes this user

Double quotes for user and password are recommended but not required. A double quote inside a user or password must be escaped with a backslash. On failure the server will reply with an I2P_ERROR and a message.

AUTH does not require that a session has been created first.

## RESULT Values

These are the values that can be carried by the RESULT field, with their meaning:

| Result | Meaning |
|--------|---------|
| OK | Operation completed successfully |
| CANT_REACH_PEER | The peer exists, but cannot be reached |
| DUPLICATED_DEST | The specified Destination is already in use |
| I2P_ERROR | A generic I2P error (e.g. I2CP disconnection, etc.) |
| INVALID_KEY | The specified key is not valid (bad format, etc.) |
| KEY_NOT_FOUND | The naming system can't resolve the given name |
| PEER_NOT_FOUND | The peer cannot be found on the network |
| TIMEOUT | Timeout while waiting for an event (e.g. peer answer) |
| LEASESET_NOT_FOUND | See Name Lookup Options above. As of router API 0.9.66 |

Different implementations may not be consistent in which RESULT is returned in various scenarios.

Most responses with a RESULT, other than OK, will also include a MESSAGE with additional information. The MESSAGE will generally be helpful in debugging problems. However, MESSAGE strings are implementation-dependent, may or may not be translated by the SAM server to the current locale, may contain internal implementation-specific information such as exceptions, and are subject to change without notice. While SAM clients may choose to expose MESSAGE strings to users, they should not make programmatic decisions based on those strings, as that will be fragile.

## Tunnel, I2CP, and Streaming Options

These options may be passed in as name=value pairs in the SAM SESSION CREATE line.

All sessions may include I2CP options such as tunnel lengths and quantities. STREAM sessions may include Streaming library options.

See those references for option names and defaults. The referenced documention is for the Java router implementation. Defaults are subject to change. Option names and values are case-sensitive. Other router implementations may not support all options and may have different defaults; consult router documentation for details.

## BASE 64 Notes

Base 64 encoding must use the I2P standard Base 64 alphabet "A-Z, a-z, 0-9, -, ~".

## Default SAM Setup

The default SAM port is 7656. SAM is not enabled by default in the Java I2P Router; it must be started manually, or configured to start automatically, on the configure clients page in the router console, or in the clients.config file. The default SAM UDP port is 7655, listening on 127.0.0.1. These may be changed in the Java router by adding the arguments sam.udp.port=nnnnn and/or sam.udp.host=w.x.y.z to the invocation, or on the SESSION line.

Configuration in other routers is implementation-specific. See the i2pd configuration guide here.