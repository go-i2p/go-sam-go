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

type Session interface {
	net.Conn
	ID() string
	Keys() i2pkeys.I2PKeys
	Close() error
	// Add other session methods as needed
}

type BaseSession struct {
	id   string
	conn net.Conn
	keys i2pkeys.I2PKeys
	SAM  SAM
}

func (bs *BaseSession) Conn() net.Conn {
	return bs.conn
}

func (bs *BaseSession) ID() string                  { return bs.id }
func (bs *BaseSession) Keys() i2pkeys.I2PKeys       { return bs.keys }
func (bs *BaseSession) Read(b []byte) (int, error)  { return bs.conn.Read(b) }
func (bs *BaseSession) Write(b []byte) (int, error) { return bs.conn.Write(b) }
func (bs *BaseSession) Close() error                { return bs.conn.Close() }

func (bs *BaseSession) LocalAddr() net.Addr {
	return bs.conn.LocalAddr()
}

func (bs *BaseSession) RemoteAddr() net.Addr {
	return bs.conn.RemoteAddr()
}

func (bs *BaseSession) SetDeadline(t time.Time) error {
	return bs.conn.SetDeadline(t)
}

func (bs *BaseSession) SetReadDeadline(t time.Time) error {
	return bs.conn.SetReadDeadline(t)
}

func (bs *BaseSession) SetWriteDeadline(t time.Time) error {
	return bs.conn.SetWriteDeadline(t)
}

func (bs *BaseSession) From() string {
	return bs.SAM.SAMEmit.I2PConfig.Fromport
}

func (bs *BaseSession) To() string {
	return bs.SAM.SAMEmit.I2PConfig.Toport
}
