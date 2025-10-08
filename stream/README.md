# go-sam-go/stream

Streaming library for TCP-like connections over I2P using the SAMv3 protocol.

## Installation

Install using Go modules with the package path `github.com/go-i2p/go-sam-go/stream`.

## Usage

The package provides TCP-like streaming connections over I2P networks. StreamSession manages the connection lifecycle, StreamListener handles incoming connections, and StreamConn implements the standard net.Conn interface.

Create sessions using NewStreamSession(), establish listeners with Listen(), and dial outbound connections using Dial() or DialI2P(). Supports context-based timeouts, concurrent operations, and automatic connection management.

## Dependencies

- github.com/go-i2p/go-sam-go/common - Core SAM protocol implementation
- github.com/go-i2p/i2pkeys - I2P cryptographic key handling
- github.com/sirupsen/logrus - Structured logging
- github.com/samber/oops - Enhanced error handling