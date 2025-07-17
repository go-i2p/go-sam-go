package raw

import (
	"github.com/go-i2p/logger"
)

// log provides the default logger instance for the raw package.
// This logger is configured to use the standard go-i2p logging system
// and provides structured logging capabilities for raw session operations.
var log = logger.GetGoI2PLogger()
