# go-sam-go/primary

Primary session library for multi-session management over I2P using the SAMv3 protocol.

## Installation

Install using Go modules with the package path `github.com/go-i2p/go-sam-go/primary`.

## Usage

The package provides primary session management for creating multiple sub-sessions over I2P networks. PrimarySession manages the primary session lifecycle and allows creation of stream, datagram, and raw sub-sessions.

Create primary sessions using NewPrimarySession(), then create sub-sessions using NewStreamSubSession(), NewDatagramSubSession(), or NewRawSubSession(). All sub-sessions share the same I2P destination.

## Dependencies

- github.com/go-i2p/go-sam-go/common - Core SAM protocol implementation
- github.com/go-i2p/i2pkeys - I2P cryptographic key handling
- github.com/sirupsen/logrus - Structured logging
- github.com/samber/oops - Enhanced error handling
