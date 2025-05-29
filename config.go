package sam3

import (
	"strconv"

	"github.com/go-i2p/go-sam-go/common"
)

// I2PConfig manages I2P configuration options
type I2PConfig struct {
	*common.I2PConfig
}

// NewConfig creates a new I2PConfig
func NewConfig(opts ...func(*I2PConfig) error) (*I2PConfig, error) {
	baseConfig, err := common.NewConfig()
	if err != nil {
		return nil, err
	}

	config := &I2PConfig{
		I2PConfig: baseConfig,
	}

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// All the configuration method forwards
func (f *I2PConfig) SetSAMAddress(addr string) {
	f.I2PConfig.SetSAMAddress(addr)
}

func (f *I2PConfig) Sam() string {
	return f.I2PConfig.Sam()
}

func (f *I2PConfig) SAMAddress() string {
	return f.I2PConfig.SAMAddress()
}

func (f *I2PConfig) ID() string {
	return f.I2PConfig.ID()
}

func (f *I2PConfig) Print() []string {
	return f.I2PConfig.Print()
}

func (f *I2PConfig) SessionStyle() string {
	return f.I2PConfig.SessionStyle()
}

func (f *I2PConfig) MinSAM() string {
	return f.I2PConfig.MinSAM()
}

func (f *I2PConfig) MaxSAM() string {
	return f.I2PConfig.MaxSAM()
}

func (f *I2PConfig) DestinationKey() string {
	return f.I2PConfig.DestinationKey()
}

func (f *I2PConfig) SignatureType() string {
	return f.I2PConfig.SignatureType()
}

func (f *I2PConfig) ToPort() string {
	return f.I2PConfig.ToPort()
}

func (f *I2PConfig) Reduce() string {
	return f.I2PConfig.Reduce()
}

func (f *I2PConfig) Reliability() string {
	return f.I2PConfig.Reliability()
}

// Configuration option setters for all the missing Set* functions
func SetInAllowZeroHop(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		e.InAllowZeroHop = s == "true"
		return nil
	}
}

func SetOutAllowZeroHop(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		e.OutAllowZeroHop = s == "true"
		return nil
	}
}

func SetInLength(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		if i, err := strconv.Atoi(s); err == nil {
			e.InLength = i
		}
		return nil
	}
}

func SetOutLength(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		if i, err := strconv.Atoi(s); err == nil {
			e.OutLength = i
		}
		return nil
	}
}

func SetInQuantity(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		if i, err := strconv.Atoi(s); err == nil {
			e.InQuantity = i
		}
		return nil
	}
}

func SetOutQuantity(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		if i, err := strconv.Atoi(s); err == nil {
			e.OutQuantity = i
		}
		return nil
	}
}

func SetInVariance(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		if i, err := strconv.Atoi(s); err == nil {
			e.InVariance = i
		}
		return nil
	}
}

func SetOutVariance(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		if i, err := strconv.Atoi(s); err == nil {
			e.OutVariance = i
		}
		return nil
	}
}

func SetInBackupQuantity(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		if i, err := strconv.Atoi(s); err == nil {
			e.InBackupQuantity = i
		}
		return nil
	}
}

func SetOutBackupQuantity(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		if i, err := strconv.Atoi(s); err == nil {
			e.OutBackupQuantity = i
		}
		return nil
	}
}

func SetUseCompression(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		e.UseCompression = s == "true"
		return nil
	}
}

func SetReduceIdleTime(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		if i, err := strconv.Atoi(s); err == nil {
			e.ReduceIdleTime = i
		}
		return nil
	}
}

func SetCloseIdleTime(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		if i, err := strconv.Atoi(s); err == nil {
			e.CloseIdleTime = i
		}
		return nil
	}
}

func SetAccessListType(s string) func(*I2PConfig) error {
	return func(e *I2PConfig) error {
		e.AccessListType = s
		return nil
	}
}
