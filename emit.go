package sam3

import (
	"github.com/go-i2p/go-sam-go/common"
)

// SAMEmit handles SAM protocol message generation
type SAMEmit struct {
	I2PConfig
	emit *common.SAMEmit
}

// NewEmit creates a new SAMEmit
func NewEmit(opts ...func(*SAMEmit) error) (*SAMEmit, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, err
	}

	emit := &SAMEmit{
		I2PConfig: *config,
		emit:      &common.SAMEmit{I2PConfig: *config.I2PConfig},
	}

	for _, opt := range opts {
		if err := opt(emit); err != nil {
			return nil, err
		}
	}

	return emit, nil
}

// Hello generates hello message
func (e *SAMEmit) Hello() string {
	return e.emit.Hello()
}

// HelloBytes generates hello message as bytes
func (e *SAMEmit) HelloBytes() []byte {
	return e.emit.HelloBytes()
}

// Create generates session create message
func (e *SAMEmit) Create() string {
	return e.emit.Create()
}

// CreateBytes generates session create message as bytes
func (e *SAMEmit) CreateBytes() []byte {
	return e.emit.CreateBytes()
}

// Connect generates connect message
func (e *SAMEmit) Connect(dest string) string {
	return e.emit.Connect(dest)
}

// ConnectBytes generates connect message as bytes
func (e *SAMEmit) ConnectBytes(dest string) []byte {
	return e.emit.ConnectBytes(dest)
}

// Accept generates accept message
func (e *SAMEmit) Accept() string {
	return e.emit.Accept()
}

// AcceptBytes generates accept message as bytes
func (e *SAMEmit) AcceptBytes() []byte {
	return e.emit.AcceptBytes()
}

// Lookup generates lookup message
func (e *SAMEmit) Lookup(name string) string {
	return e.emit.Lookup(name)
}

// LookupBytes generates lookup message as bytes
func (e *SAMEmit) LookupBytes(name string) []byte {
	return e.emit.LookupBytes(name)
}

// GenerateDestination generates destination message
func (e *SAMEmit) GenerateDestination() string {
	return e.emit.GenerateDestination()
}

// GenerateDestinationBytes generates destination message as bytes
func (e *SAMEmit) GenerateDestinationBytes() []byte {
	return e.emit.GenerateDestinationBytes()
}

// SamOptionsString returns SAM options as string
func (e *SAMEmit) SamOptionsString() string {
	return e.emit.SamOptionsString()
}
