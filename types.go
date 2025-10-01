// Package sam provides type aliases and wrappers for exposing sub-package types at the root level.
// This file implements the sam3-compatible API surface by creating type aliases that delegate
// to the appropriate sub-package implementations while maintaining a clean public interface.
package sam

import (
	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/go-sam-go/datagram"
	"github.com/go-i2p/go-sam-go/raw"
	"github.com/go-i2p/go-sam-go/stream"
)

// Core SAM types - These provide the fundamental SAM bridge functionality
// and configuration management that applications use to connect to I2P.

// SAM represents the core SAM bridge connection and provides methods for session creation
// and I2P address resolution. It delegates to the common.SAM implementation while
// exposing the sam3-compatible interface at the root package level.
type SAM = common.SAM

// SAMResolver provides I2P address resolution services through the SAM protocol.
// It wraps the common.SAMResolver to provide name-to-address lookup functionality
// that applications need for connecting to I2P destinations by name.
type SAMResolver = common.SAMResolver

// Configuration types - These manage I2P tunnel parameters and session options
// that control anonymity, performance, and network behavior characteristics.

// I2PConfig manages I2P tunnel configuration options including tunnel length,
// backup quantities, variance settings, and other I2P-specific parameters.
// Applications use this to customize their anonymity and performance characteristics.
type I2PConfig = common.I2PConfig

// SAMEmit handles SAM protocol message generation and configuration management.
// It embeds I2PConfig to provide comprehensive session configuration capabilities
// while managing the underlying SAM protocol communication requirements.
type SAMEmit = common.SAMEmit

// Options represents a map of configuration options that can be applied to sessions.
// This provides a flexible way to specify tunnel parameters and session behaviors
// using key-value pairs that are converted to SAM protocol commands.
type Options = common.Options

// Option represents a functional option for configuring SAMEmit instances.
// This follows the functional options pattern to provide type-safe configuration
// with clear error handling and composable session parameter management.
type Option = common.Option

// Session types - These provide different communication patterns for I2P applications
// including streaming (TCP-like), datagram (UDP-like), and raw (unrepliable) modes.

// StreamSession provides TCP-like reliable connection capabilities over I2P networks.
// It supports both client and server operations with connection multiplexing,
// listener management, and standard Go networking interfaces for streaming data.
type StreamSession = stream.StreamSession

// DatagramSession provides UDP-like messaging capabilities with message reliability
// and ordering guarantees over I2P networks. It handles signed, authenticated
// messaging with replay protection for secure datagram communication patterns.
type DatagramSession = datagram.DatagramSession

// RawSession provides unrepliable datagram communication over I2P networks.
// Messages are encrypted end-to-end but senders cannot be identified or replied to,
// providing the highest level of sender anonymity for one-way communication.
type RawSession = raw.RawSession

// Connection types - These implement standard Go networking interfaces
// for seamless integration with existing networking code and patterns.

// SAMConn implements net.Conn for I2P streaming connections, providing a standard
// Go networking interface for TCP-like reliable communication over I2P networks.
// This alias maintains sam3 API compatibility while delegating to stream.StreamConn.
type SAMConn = stream.StreamConn

// StreamListener implements net.Listener for I2P streaming connections.
// It manages incoming connection acceptance and provides thread-safe operations
// for accepting connections from remote I2P destinations in server applications.
type StreamListener = stream.StreamListener

// PrimarySession provides master session capabilities that can create and manage
// multiple sub-sessions of different types (stream, datagram, raw) within a single
// I2P session context. This enables complex applications with multiple communication
// patterns while sharing the same I2P identity and tunnel infrastructure.
//
// Note: This is a placeholder type that will be implemented when the primary
// package is fully developed. Currently returns a basic structure that will
// be enhanced with full sub-session management capabilities.
type PrimarySession struct {
	// sam holds the underlying SAM connection for session management
	sam *SAM

	// id uniquely identifies this primary session within the SAM bridge
	id string

	// options contains the configuration parameters for this session
	options []string

	// TODO: Add sub-session management when primary package is implemented
	// This will include methods for creating stream, datagram, and raw sub-sessions
	// with proper lifecycle management and resource cleanup capabilities.
}

// BaseSession represents the underlying session functionality that all session types
// extend. It provides common operations like connection management, key access,
// and standard net.Conn interface implementation for I2P session operations.
type BaseSession = common.BaseSession
