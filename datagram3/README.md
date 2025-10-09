# go-sam-go/datagram3

Datagram3 library for hash-based UDP-like messaging with hash-based sources over I2P using the SAMv3 protocol.

## Installation

Install using Go modules with the package path `github.com/go-i2p/go-sam-go/datagram3`.

## Usage

The package provides hash-based datagram messaging over I2P networks. Datagram3Session manages the session lifecycle, Datagram3Reader handles incoming datagrams, Datagram3Writer sends outgoing datagrams, and Datagram3Conn implements the standard net.PacketConn interface.

Create sessions using NewDatagram3Session(), send messages with SendDatagram3(), and receive messages using ReceiveDatagram3(). Sources use 32-byte hashes that are not cryptographically authenticated. Implement application-layer authentication for security. Requires I2P router with DATAGRAM3 support.

## Dependencies

- github.com/go-i2p/go-sam-go/common - Core SAM protocol implementation
- github.com/go-i2p/i2pkeys - I2P cryptographic key handling
- github.com/sirupsen/logrus - Structured logging
- github.com/samber/oops - Enhanced error handling
