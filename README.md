# go-sam-go

[![Go Reference](https://pkg.go.dev/badge/github.com/go-i2p/go-sam-go.svg)](https://pkg.go.dev/github.com/go-i2p/go-sam-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-i2p/go-sam-go)](https://goreportcard.com/report/github.com/go-i2p/go-sam-go)

A pure-Go implementation of SAMv3.3 (Simple Anonymous Messaging) for I2P, focused on maintainability and clean architecture. This project is forked from `github.com/go-i2p/sam3` with reorganized code structure.

**WARNING: This is a new package. Streaming works. Repliable datagrams and Raw datagrams work. Primary Sessions, work but are untested. Authenticated Repliable Datagrams(Datagram2), and Unauthenticated Repliable Datagrams(Datagram3) are NOT YET IMPLEMENTED.**
**The API should not change much.**
**It needs more people looking at it.**

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

#### `datagram2` Package (Planned - Not Implemented)
Authenticated repliable datagrams:
```go
// Will be available in future release - currently not implemented
// dgram2, err := sam.NewDatagram2Session("udp", keys, options, 0)
// n, err := dgram2.WriteTo(data, dest)
```

#### `datagram3` Package (Planned - Not Implemented)
Unauthenticated repliable datagrams:
```go
// Will be available in future release - currently not implemented
// dgram3, err := sam.NewDatagram3Session("udp", keys, options, 0)
// n, err := dgram3.WriteTo(data, dest)
```

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
- Running I2P router with SAM enabled (default port: 7656)

## 📝 Development

```bash
# Format code
make fmt

# Run tests
go test ./...
```

## 📄 License

MIT License

## 🙏 Acknowledgments

Based on the original [github.com/go-i2p/sam3](https://github.com/go-i2p/sam3) library.
