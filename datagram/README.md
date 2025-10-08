# go-sam-go/datagram

Datagram library for authenticated UDP-like message delivery over I2P using the SAMv3 protocol.

## Installation

Install using Go modules with the package path `github.com/go-i2p/go-sam-go/datagram`.

## Usage

The package provides authenticated datagram messaging over I2P networks. DatagramSession manages the session lifecycle, DatagramReader handles incoming datagrams, DatagramWriter sends outgoing datagrams, and DatagramConn implements the standard net.PacketConn interface.

Create sessions using NewDatagramSession(), send messages with SendDatagram(), and receive messages using ReceiveDatagram(). Supports I2P address resolution, configurable tunnel parameters, and proper resource cleanup.

## Dependencies

- github.com/go-i2p/go-sam-go/common - Core SAM protocol implementation
- github.com/go-i2p/i2pkeys - I2P cryptographic key handling
- github.com/sirupsen/logrus - Structured logging
- github.com/samber/oops - Enhanced error handling
