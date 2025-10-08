// Package common provides core SAM protocol implementation and shared utilities for I2P.
//
// This package implements the foundational SAMv3.3 protocol communication layer, including
// session management, configuration options, I2P name resolution, and shared abstractions
// used by all session types (stream, datagram, raw).
//
// Core types:
//   - SAM: Base connection to I2P SAM bridge (default port 7656)
//   - Session: Base session interface with lifecycle management
//   - I2PConfig: Configuration builder for tunnel parameters
//   - SAMEmit: SAM protocol command formatter
//
// Session creation requires 2-5 minutes for I2P tunnel establishment; use generous timeouts
// and exponential backoff retry logic. All network operations should use context.Context
// for cancellation and timeout control.
//
// Basic usage:
//
//	sam, err := common.NewSAM("127.0.0.1:7656")
//	if err != nil { log.Fatal(err) }
//	defer sam.Close()
//	session, err := sam.NewGenericSession("STREAM", "my-session", keys, []string{"inbound.length=1"})
//	defer session.Close()
//
// This package is primarily used as a foundation by higher-level packages (stream, datagram,
// raw, primary) rather than being used directly by applications.
package common
