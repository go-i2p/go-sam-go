package common

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-i2p/i2pkeys"
)

// I2PConfig is a struct which manages I2P configuration options.
type I2PConfig struct {
	SamHost string
	SamPort int
	TunName string

	SamMin string
	SamMax string

	Fromport string
	Toport   string

	Style   string
	TunType string

	DestinationKeys *i2pkeys.I2PKeys

	SigType                   string
	EncryptLeaseSet           bool
	LeaseSetKey               string
	LeaseSetPrivateKey        string
	LeaseSetPrivateSigningKey string
	LeaseSetKeys              i2pkeys.I2PKeys
	InAllowZeroHop            bool
	OutAllowZeroHop           bool
	InLength                  int
	OutLength                 int
	InQuantity                int
	OutQuantity               int
	InVariance                int
	OutVariance               int
	InBackupQuantity          int
	OutBackupQuantity         int
	FastRecieve               bool
	UseCompression            bool
	MessageReliability        string
	CloseIdle                 bool
	CloseIdleTime             int
	ReduceIdle                bool
	ReduceIdleTime            int
	ReduceIdleQuantity        int
	LeaseSetEncryption        string

	// Streaming Library options
	AccessListType string
	AccessList     []string
}

type SAMEmit struct {
	I2PConfig
}

// Used for controlling I2Ps SAMv3.
type SAM struct {
	SAMEmit
	SAMResolver
	net.Conn

	// Timeout for SAM connections
	Timeout time.Duration
	// Context for control of lifecycle
	Context context.Context
}

type SAMResolver struct {
	*SAM
}

// options map
type Options map[string]string

// obtain sam options as list of strings
func (opts Options) AsList() (ls []string) {
	for k, v := range opts {
		ls = append(ls, fmt.Sprintf("%s=%s", k, v))
	}
	return
}
