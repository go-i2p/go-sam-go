# go-sam-go/common

Core library for SAMv3 protocol implementation in Go, providing connection management and session configuration for I2P applications.

## Installation

Install using Go modules with the package path `github.com/go-i2p/go-sam-go/common`.

## Usage

The package handles SAM bridge connections, handshakes, and base session management. It provides configuration options for tunnel parameters, encryption settings, and I2P-specific features. The BaseSession implementation must be wrapped in specific session types (stream, datagram, or raw) for actual use.

Key components include SAM connection establishment, I2P address resolution, destination key management, and comprehensive tunnel configuration through the I2PConfig struct.

## Dependencies

- github.com/go-i2p/i2pkeys - I2P cryptographic key handling
- github.com/go-i2p/logger - Logging functionality
- github.com/sirupsen/logrus - Structured logging
- github.com/samber/oops - Enhanced error handling