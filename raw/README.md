# go-sam-go/raw

Raw datagram library for encrypted but unauthenticated message delivery over I2P using the SAMv3 protocol.

## Installation

Install using Go modules with the package path `github.com/go-i2p/go-sam-go/raw`.

## Usage

The package provides encrypted but unauthenticated datagram messaging over I2P networks. RawSession manages the session lifecycle, RawReader handles incoming datagrams, RawWriter sends outgoing datagrams, and RawConn implements the standard net.PacketConn interface.

Create sessions using NewRawSession(), send messages with SendDatagram(), and receive messages using ReceiveDatagram(). Messages are encrypted but senders are not authenticated. Implement application-layer authentication for security.

## Dependencies

- github.com/go-i2p/go-sam-go/common - Core SAM protocol implementation
- github.com/go-i2p/i2pkeys - I2P cryptographic key handling
- github.com/sirupsen/logrus - Structured logging
- github.com/samber/oops - Enhanced error handling
