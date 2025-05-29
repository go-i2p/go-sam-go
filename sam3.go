// Package sam3 provides a compatibility layer for the go-i2p/sam3 library using go-sam-go as the backend
package sam3

import (
	"io"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// Constants from original sam3
const (
	Sig_NONE                 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"
	Sig_DSA_SHA1             = "SIGNATURE_TYPE=DSA_SHA1"
	Sig_ECDSA_SHA256_P256    = "SIGNATURE_TYPE=ECDSA_SHA256_P256"
	Sig_ECDSA_SHA384_P384    = "SIGNATURE_TYPE=ECDSA_SHA384_P384"
	Sig_ECDSA_SHA512_P521    = "SIGNATURE_TYPE=ECDSA_SHA512_P521"
	Sig_EdDSA_SHA512_Ed25519 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"
)

// Predefined option sets (keeping your existing definitions)
var (
	Options_Humongous = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=3", "outbound.backupQuantity=3",
		"inbound.quantity=6", "outbound.quantity=6",
	}

	Options_Large = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=1", "outbound.backupQuantity=1",
		"inbound.quantity=4", "outbound.quantity=4",
	}

	Options_Wide = []string{
		"inbound.length=1", "outbound.length=1",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=2", "outbound.backupQuantity=2",
		"inbound.quantity=3", "outbound.quantity=3",
	}

	Options_Medium = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=2", "outbound.quantity=2",
	}

	Options_Default = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=0", "outbound.lengthVariance=0",
		"inbound.backupQuantity=1", "outbound.backupQuantity=1",
		"inbound.quantity=1", "outbound.quantity=1",
	}

	Options_Small = []string{
		"inbound.length=3", "outbound.length=3",
		"inbound.lengthVariance=1", "outbound.lengthVariance=1",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=1", "outbound.quantity=1",
	}

	Options_Warning_ZeroHop = []string{
		"inbound.length=0", "outbound.length=0",
		"inbound.lengthVariance=0", "outbound.lengthVariance=0",
		"inbound.backupQuantity=0", "outbound.backupQuantity=0",
		"inbound.quantity=2", "outbound.quantity=2",
	}
)

// Global variables from original sam3
var (
	PrimarySessionSwitch string = PrimarySessionString()
	SAM_HOST                    = getEnv("sam_host", "127.0.0.1")
	SAM_PORT                    = getEnv("sam_port", "7656")
)

// SAM represents the main controller for I2P router's SAM bridge
type SAM struct {
	Config SAMEmit
	sam    *common.SAM
}

// NewSAM creates a new controller for the I2P routers SAM bridge
func NewSAM(address string) (*SAM, error) {
	samInstance, err := common.NewSAM(address)
	if err != nil {
		return nil, err
	}

	config, err := NewConfig()
	if err != nil {
		return nil, err
	}

	emit := SAMEmit{I2PConfig: *config}

	return &SAM{
		Config: emit,
		sam:    samInstance,
	}, nil
}

// Close closes this sam session
func (sam *SAM) Close() error {
	return sam.sam.Close()
}

// Keys returns the keys associated with this SAM instance
func (sam *SAM) Keys() (k *i2pkeys.I2PKeys) {
	return sam.sam.Keys()
}

// NewKeys creates the I2P-equivalent of an IP address
func (sam *SAM) NewKeys(sigType ...string) (i2pkeys.I2PKeys, error) {
	return sam.sam.NewKeys(sigType...)
}

// ReadKeys reads public/private keys from an io.Reader
func (sam *SAM) ReadKeys(r io.Reader) (err error) {
	return sam.sam.ReadKeys(r)
}

// EnsureKeyfile ensures keyfile exists
func (sam *SAM) EnsureKeyfile(fname string) (keys i2pkeys.I2PKeys, err error) {
	return sam.sam.EnsureKeyfile(fname)
}

// Lookup performs a name lookup
func (sam *SAM) Lookup(name string) (i2pkeys.I2PAddr, error) {
	return sam.sam.Lookup(name)
}
