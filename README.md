# go-sam-go

[![Go Reference](https://pkg.go.dev/badge/github.com/go-i2p/go-sam-go.svg)](https://pkg.go.dev/github.com/go-i2p/go-sam-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-i2p/go-sam-go)](https://goreportcard.com/report/github.com/go-i2p/go-sam-go)

A pure-Go implementation of SAMv3.3 (Simple Anonymous Messaging) for I2P, focused on maintainability and clean architecture. This project is forked from `github.com/go-i2p/sam3` with reorganized code structure.

## ⚠️ Implementation Status

**Stable & Production-Ready:**
- ✅ **Stream** - TCP-like reliable connections (fully tested)
- ✅ **Datagram** - Legacy authenticated repliable datagrams (fully tested)
- ✅ **Raw** - Encrypted unauthenticated datagrams (fully tested)

**Implemented & Documented (Awaiting I2P Router Support):**
- ⚠️ **Datagram2** - Authenticated repliable datagrams with replay protection (spec finalized early 2025, no router implementations yet)
- ⚠️ **Datagram3** - Unauthenticated repliable datagrams with hash-based sources (spec finalized early 2025, no router implementations yet)

**Partially Implemented:**
- 🔶 **Primary Sessions** - Multi-session management (works but needs more testing)

**Note:** DATAGRAM2 and DATAGRAM3 are fully implemented in this library but require I2P router support (Java I2P or i2pd) to function. Check your router's release notes for SAMv3 DATAGRAM2/3 support.

## 📦 Installation

```bash
go get github.com/go-i2p/go-sam-go
```

### Dependencies

- `github.com/go-i2p/i2pkeys` v0.33.92 - I2P cryptographic key management
- `github.com/go-i2p/crypto` - I2P-specific cryptographic operations
- `github.com/sirupsen/logrus` v1.9.3 - Structured logging
- `github.com/samber/oops` v1.19.0 - Enhanced error handling

## 🚀 Quick Start

```go
package main

import (
    "github.com/go-i2p/go-sam-go"
)

func main() {
    // Create SAM client
    client, err := sam3.NewSAM("127.0.0.1:7656")
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    // Generate keys (optionally specify signature type)
    keys, err := client.NewKeys() // Uses default EdDSA_SHA512_Ed25519
    // Or: keys, err := client.NewKeys(sam3.Sig_ECDSA_SHA256_P256)
    if err != nil {
        panic(err)
    }
    
    // Create streaming session
    session, err := client.NewStreamSession("myTunnel", keys, sam3.Options_Default)
    if err != nil {
        panic(err)
    }
}
```

## 📚 API Documentation

### Root Package (`sam3`)
The root package provides a high-level wrapper API:

```go
client, err := sam3.NewSAM("127.0.0.1:7656")
```

Available session types:
- `NewStreamSession()` - For reliable TCP-like connections
- `NewDatagramSession()` - For UDP-like messaging 
- `NewRawSession()` - For encrypted but unauthenticated datagrams
- `NewPrimarySession()` - For creating multiple sub-sessions

### Sub-packages

#### `primary` Package
Core session management functionality:
```go
primary, err := sam.NewPrimarySession("mainSession", keys, options)
sub1, err := primary.NewStreamSubSession("web")
sub2, err := primary.NewDatagramSubSession("chat") 
```

#### `stream` Package 
TCP-like reliable connections:
```go
listener, err := session.Listen()
conn, err := session.Accept()
// or
conn, err := session.DialI2P(remote)
```

#### `datagram` Package
UDP-like message delivery:
```go
dgram, err := sam.NewDatagramSession("udp", keys, options, 0) // 0 = use default UDP port
n, err := dgram.WriteTo(data, dest)
```

#### `raw` Package
Low-level datagram access:
```go
raw, err := sam.NewRawSession("raw", keys, options, 0) // 0 = auto-assign UDP port
n, err := raw.WriteTo(data, dest)
```

#### `datagram2` Package
Authenticated repliable datagrams with replay protection:
```go
// DATAGRAM2 - Authenticated with replay protection (requires router support)
// Specification finalized early 2025, awaiting I2P router implementation
import "github.com/go-i2p/go-sam-go/datagram2"

session, err := datagram2.NewDatagram2Session(sam, "session-id", keys, options)
if err != nil {
    // Handle error - may fail if router doesn't support DATAGRAM2 yet
    log.Fatal(err)
}
defer session.Close()

// Send authenticated datagram
err = session.SendDatagram(data, destination)

// Receive with full authentication
conn := session.PacketConn()
n, addr, err := conn.ReadFrom(buffer)
```

**Security:** Provides cryptographic authentication and replay protection. Recommended for new applications requiring source verification.

**Status:** Implementation complete. Waiting for I2P router support (Java I2P 0.9.x+ or i2pd 2.x+).

#### `datagram3` Package
⚠️ **SECURITY WARNING:** Unauthenticated repliable datagrams with hash-based sources:
```go
// DATAGRAM3 - UNAUTHENTICATED sources (requires router support + app-layer auth)
// Sources are 32-byte hashes that can be spoofed - implement your own authentication!
import "github.com/go-i2p/go-sam-go/datagram3"

session, err := datagram3.NewDatagram3Session(sam, "session-id", keys, options)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

// Receive datagram with UNAUTHENTICATED source hash
reader := session.NewReader()
datagram, err := reader.ReceiveDatagram()

// ⚠️ Source hash is NOT authenticated - verify at application layer!
// Resolve hash to destination for reply (requires NAMING LOOKUP)
err = datagram.ResolveSource(session)

// Send reply
writer := session.NewWriter()
err = writer.SendDatagram(replyData, datagram.Source)
```

**Security:** ⚠️ **Sources are NOT authenticated and can be spoofed!** Only use when:
- Application implements message-level authentication (signatures, HMAC, etc.)
- Source identity is not security-critical
- Lower overhead is essential

**Use DATAGRAM2 instead if you need authenticated sources.**

**Status:** Implementation complete with comprehensive security documentation. Waiting for I2P router support.

### Configuration

Built-in configuration profiles:
```go
sam3.Options_Default     // Balanced defaults
sam3.Options_Small      // Minimal resources
sam3.Options_Medium     // Enhanced reliability 
sam3.Options_Large      // High throughput
sam3.Options_Humongous  // Maximum performance
```

Debug logging:
```bash
export DEBUG_I2P=debug   # Debug level
export DEBUG_I2P=warn    # Warning level
export DEBUG_I2P=error   # Error level
```

## 🔧 Requirements

- Go 1.24.2 or later (toolchain go1.24.4)
- Running I2P router with SAM bridge enabled (default port: 7656)
- For DATAGRAM2/DATAGRAM3: I2P router with SAMv3 DATAGRAM2/3 support (check router release notes)

## 📝 Development

```bash
# Format code
make fmt

# Run tests (short mode, no I2P required)
go test -short ./...

# Run full integration tests (requires running I2P router)
# Note: I2P tests can take 30-150 seconds due to tunnel establishment
go test ./...

# Run with race detection
go test -race -short ./...
```

## 📖 Package Documentation

Each sub-package has comprehensive documentation:

- **[datagram2/](datagram2/README.md)** - DATAGRAM2 authenticated datagrams with replay protection
- **[datagram3/](datagram3/README.md)** - ⚠️ DATAGRAM3 unauthenticated datagrams (security warnings!)
- **[stream/](stream/)** - TCP-like reliable connections
- **[datagram/](datagram/)** - Legacy authenticated datagrams
- **[raw/](raw/)** - Encrypted unauthenticated datagrams
- **[primary/](primary/)** - PRIMARY session management

## 🔒 Security Notes

### DATAGRAM3 Security Warning

⚠️ **CRITICAL:** DATAGRAM3 sources are **NOT authenticated**. Any attacker can claim to be any sender by providing a fake hash. Only use DATAGRAM3 when:

1. You implement application-layer authentication (Ed25519 signatures, HMAC, etc.)
2. Source identity is not security-critical
3. You understand the security implications

**For authenticated sources, use DATAGRAM2 instead.**

See [datagram3/AUDIT.md](datagram3/AUDIT.md) for comprehensive security analysis including attack scenarios and mitigations.

### I2P Timing Considerations

I2P operations have significant latency due to tunnel-based architecture:

- **Session creation:** 2-5 minutes on initial connection
- **Message delivery:** Variable (network-dependent)
- **Best practice:** Use generous timeouts (5+ minutes) and exponential backoff retry logic

All tests accommodate I2P timing requirements.

## 📄 License

MIT License

## 🙏 Acknowledgments

Based on the original [github.com/go-i2p/sam3](https://github.com/go-i2p/sam3) library.
