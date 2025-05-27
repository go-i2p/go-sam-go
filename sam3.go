// package sam3 wraps the original sam3 API from github.com/go-i2p/sam3
package sam3

import (
	"math/rand"

	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/go-sam-go/datagram"
	"github.com/go-i2p/go-sam-go/primary"
	"github.com/go-i2p/go-sam-go/stream"
	"github.com/go-i2p/i2pkeys"
)

// Used for controlling I2Ps SAMv3.
type SAM struct {
	*common.SAM
}

// Creates a new stream session by wrapping stream.NewStreamSession
func (s *SAM) NewStreamSession(param1 string, keys i2pkeys.I2PKeys, param3 []string) (*StreamSession, error) {
	sam := &stream.SAM{
		SAM: s.SAM,
	}
	ss, err := sam.NewStreamSession(param1, keys, param3)
	if err != nil {
		return nil, err
	}
	streamSession := &StreamSession{
		StreamSession: ss,
	}
	return streamSession, nil
}

// Creates a new Datagram session by wrapping datagram.NewDatagramSession
func (s *SAM) NewDatagramSession(id string, keys i2pkeys.I2PKeys, options []string, port int) (*DatagramSession, error) {
	sam := datagram.SAM(*s.SAM)
	dgs, err := sam.NewDatagramSession(id, keys, options, port)
	if err != nil {
		return nil, err
	}
	datagramSession := DatagramSession{
		DatagramSession: *dgs,
	}
	return &datagramSession, nil
}

func (s *SAM) NewPrimarySession(id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error) {
	sam := primary.SAM(*s.SAM)
	ps, err := sam.NewPrimarySession(id, keys, options)
	if err != nil {
		return nil, err
	}
	primarySession := PrimarySession{
		PrimarySession: ps,
	}
	return &primarySession, nil
}

const (
	session_OK             = "SESSION STATUS RESULT=OK DESTINATION="
	session_DUPLICATE_ID   = "SESSION STATUS RESULT=DUPLICATED_ID\n"
	session_DUPLICATE_DEST = "SESSION STATUS RESULT=DUPLICATED_DEST\n"
	session_INVALID_KEY    = "SESSION STATUS RESULT=INVALID_KEY\n"
	session_I2P_ERROR      = "SESSION STATUS RESULT=I2P_ERROR MESSAGE="
)

const (
	Sig_NONE                 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"
	Sig_DSA_SHA1             = "SIGNATURE_TYPE=DSA_SHA1"
	Sig_ECDSA_SHA256_P256    = "SIGNATURE_TYPE=ECDSA_SHA256_P256"
	Sig_ECDSA_SHA384_P384    = "SIGNATURE_TYPE=ECDSA_SHA384_P384"
	Sig_ECDSA_SHA512_P521    = "SIGNATURE_TYPE=ECDSA_SHA512_P521"
	Sig_EdDSA_SHA512_Ed25519 = "SIGNATURE_TYPE=EdDSA_SHA512_Ed25519"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString() string {
	n := 4
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	log.WithField("randomString", string(b)).Debug("Generated random string")
	return string(b)
}

// Creates a new controller for the I2P routers SAM bridge.
func NewSAM(address string) (*SAM, error) {
	is, err := common.NewSAM(address)
	if err != nil {
		log.WithError(err).Error("Failed to create new SAM instance")
		return nil, err
	}
	s := &SAM{
		SAM: is,
	}
	return s, nil
}
