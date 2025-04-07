// package sam3 wraps the original sam3 API from github.com/go-i2p/sam3
package sam3

import "github.com/go-i2p/go-sam-go/logger"

var log = logger.GetSAM3Logger()

func init() {
	logger.InitializeSAM3Logger()
	log = logger.GetSAM3Logger()
}
