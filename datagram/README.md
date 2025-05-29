package datagram

High-level datagram library for UDP-like message delivery over I2P using the SAMv3 protocol.

## Installation

Install using Go modules with the package path `github.com/go-i2p/go-sam-go/datagram`.

## Usage

The package provides UDP-like datagram messaging over I2P networks. [`DatagramSession`](datagram/types.go) manages the session lifecycle, [`DatagramReader`](datagram/types.go) handles incoming datagrams, [`DatagramWriter`](datagram/types.go) sends outgoing datagrams, and [`DatagramConn`](datagram/types.go) implements the standard `net.PacketConn` interface for seamless integration with existing Go networking code.

Create sessions using [`NewDatagramSession`](datagram/session.go), send messages with [`SendDatagram()`](datagram/session.go), and receive messages using [`ReceiveDatagram()`](datagram/session.go). The implementation supports I2P address resolution, configurable tunnel parameters, and comprehensive error handling with proper resource cleanup.

Key features include full `net.PacketConn` and `net.Conn` compatibility, I2P destination management, base64 payload encoding, and concurrent datagram processing with proper synchronization.

## Dependencies

- github.com/go-i2p/go-sam-go/common - Core SAM protocol implementation
- github.com/go-i2p/i2pkeys - I2P cryptographic key handling  
- github.com/go-i2p/logger - Logging functionality
- github.com/sirupsen/logrus - Structured logging
- github.com/samber/oops - Enhanced error handling