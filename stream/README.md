# go-sam-go/stream

High-level streaming library for reliable TCP-like connections over I2P using the SAMv3 protocol.

## Installation

Install using Go modules with the package path `github.com/go-i2p/go-sam-go/stream`.

## Usage

The package provides TCP-like streaming connections over I2P networks. [`StreamSession`](stream/types.go) manages the connection lifecycle, [`StreamListener`](stream/types.go) handles incoming connections, and [`StreamConn`](stream/types.go) implements the standard `net.Conn` interface for seamless integration with existing Go networking code.

Create sessions using [`NewStreamSession`](stream/session.go), establish listeners with [`Listen()`](stream/session.go), and dial outbound connections using [`Dial()`](stream/session.go) or [`DialI2P()`](stream/session.go). The implementation supports context-based timeouts, concurrent operations, and automatic connection management.

Key features include full `net.Listener` and `net.Conn` compatibility, I2P address resolution, configurable tunnel parameters, and comprehensive error handling with proper resource cleanup.

## Dependencies

- github.com/go-i2p/go-sam-go/common - Core SAM protocol implementation
- github.com/go-i2p/i2pkeys - I2P cryptographic key handling  
- github.com/go-i2p/logger - Logging functionality
- github.com/sirupsen/logrus - Structured logging
- github.com/samber/oops - Enhanced error handling