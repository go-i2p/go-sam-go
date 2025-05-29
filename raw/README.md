# go-sam-go/raw

High-level raw datagram library for unencrypted message delivery over I2P using the SAMv3 protocol.

## Installation

Install using Go modules with the package path `github.com/go-i2p/go-sam-go/raw`.

## Usage

The package provides unencrypted raw datagram messaging over I2P networks. [`RawSession`](raw/types.go) manages the session lifecycle, [`RawReader`](raw/types.go) handles incoming raw datagrams, [`RawWriter`](raw/types.go) sends outgoing raw datagrams, and [`RawConn`](raw/types.go) implements the standard `net.PacketConn` interface for seamless integration with existing Go networking code.

Create sessions using [`NewRawSession`](raw/session.go), send messages with [`SendDatagram()`](raw/session.go), and receive messages using [`ReceiveDatagram()`](raw/session.go). The implementation supports I2P address resolution, configurable tunnel parameters, and comprehensive error handling with proper resource cleanup.

Key features include full `net.PacketConn` and `net.Conn` compatibility, I2P destination management, base64 payload encoding, and concurrent raw datagram processing with proper synchronization.

## Dependencies

- github.com/go-i2p/go-sam-go/common - Core SAM protocol implementation
- github.com/go-i2p/i2pkeys - I2P cryptographic key handling  
- github.com/go-i2p/logger - Logging functionality
- github.com/sirupsen/logrus - Structured logging
- github.com/samber/oops - Enhanced error handling