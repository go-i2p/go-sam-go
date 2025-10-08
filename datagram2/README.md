# go-sam-go/datagram2

Datagram2 library for authenticated UDP-like messaging with replay protection over I2P using the SAMv3 protocol.

## Installation

Install using Go modules with the package path `github.com/go-i2p/go-sam-go/datagram2`.

## Usage

The package provides authenticated datagram messaging with replay protection over I2P networks. Datagram2Session manages the session lifecycle, Datagram2Reader handles incoming datagrams, Datagram2Writer sends outgoing datagrams, and Datagram2Conn implements the standard net.PacketConn interface.

Create sessions using NewDatagram2Session(), send messages with SendDatagram2(), and receive messages using ReceiveDatagram2(). Requires I2P router with DATAGRAM2 support. Check router release notes for compatibility.

## Dependencies

- github.com/go-i2p/go-sam-go/common - Core SAM protocol implementation
- github.com/go-i2p/i2pkeys - I2P cryptographic key handling
- github.com/sirupsen/logrus - Structured logging
- github.com/samber/oops - Enhanced error handling
