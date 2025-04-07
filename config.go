// package sam3 wraps the original sam3 API from github.com/go-i2p/sam3
package sam3

import (
	"fmt"

	"github.com/go-i2p/go-sam-go/common"
)

// I2PConfig is a struct which manages I2P configuration options
type I2PConfig struct {
	common.I2PConfig
}

func NewConfig(opts ...func(*I2PConfig) error) (*I2PConfig, error) {
	var config I2PConfig
	config.SamHost = "127.0.0.1"
	config.SamPort = 7656
	config.SamMin = common.DEFAULT_SAM_MIN
	config.SamMax = common.DEFAULT_SAM_MAX
	config.TunName = ""
	config.TunType = "server"
	config.Style = common.SESSION_STYLE_STREAM
	config.InLength = 3
	config.OutLength = 3
	config.InQuantity = 2
	config.OutQuantity = 2
	config.InVariance = 1
	config.OutVariance = 1
	config.InBackupQuantity = 3
	config.OutBackupQuantity = 3
	config.InAllowZeroHop = false
	config.OutAllowZeroHop = false
	config.EncryptLeaseSet = false
	config.LeaseSetKey = ""
	config.LeaseSetPrivateKey = ""
	config.LeaseSetPrivateSigningKey = ""
	config.FastRecieve = false
	config.UseCompression = true
	config.ReduceIdle = false
	config.ReduceIdleTime = 15
	config.ReduceIdleQuantity = 4
	config.CloseIdle = false
	config.CloseIdleTime = 300000
	config.MessageReliability = "none"
	for _, o := range opts {
		if err := o(&config); err != nil {
			return nil, err
		}
	}
	return &config, nil
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
