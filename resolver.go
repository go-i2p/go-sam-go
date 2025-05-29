package sam3

import (
	"github.com/go-i2p/go-sam-go/common"
	"github.com/go-i2p/i2pkeys"
)

// SAMResolver provides name resolution functionality
type SAMResolver struct {
	*SAM
	resolver *common.SAMResolver
}

// NewSAMResolver creates a new SAMResolver from existing SAM
func NewSAMResolver(parent *SAM) (*SAMResolver, error) {
	resolver, err := common.NewSAMResolver(parent.sam)
	if err != nil {
		return nil, err
	}

	return &SAMResolver{
		SAM:      parent,
		resolver: resolver,
	}, nil
}

// NewFullSAMResolver creates a new full SAMResolver
func NewFullSAMResolver(address string) (*SAMResolver, error) {
	resolver, err := common.NewFullSAMResolver(address)
	if err != nil {
		return nil, err
	}

	sam := &SAM{sam: resolver.SAM}
	return &SAMResolver{
		SAM:      sam,
		resolver: resolver,
	}, nil
}

// Resolve performs a lookup
func (sam *SAMResolver) Resolve(name string) (i2pkeys.I2PAddr, error) {
	return sam.resolver.Resolve(name)
}
