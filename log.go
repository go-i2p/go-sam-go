package sam3

<<<<<<< HEAD
import (
	"github.com/go-i2p/logger"
)

var log = logger.GetGoI2PLogger()
||||||| a681ab3
import (
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	log  *logrus.Logger
	once sync.Once
)

func InitializeSAM3Logger() {
	once.Do(func() {
		log = logrus.New()
		// We do not want to log by default
		log.SetOutput(ioutil.Discard)
		log.SetLevel(logrus.PanicLevel)
		// Check if DEBUG_I2P is set
		if logLevel := os.Getenv("DEBUG_I2P"); logLevel != "" {
			log.SetOutput(os.Stdout)
			switch strings.ToLower(logLevel) {
			case "debug":
				log.SetLevel(logrus.DebugLevel)
			case "warn":
				log.SetLevel(logrus.WarnLevel)
			case "error":
				log.SetLevel(logrus.ErrorLevel)
			default:
				log.SetLevel(logrus.DebugLevel)
			}
			log.WithField("level", log.GetLevel()).Debug("Logging enabled.")
		}
	})
}

// GetSAM3Logger returns the initialized logger
func GetSAM3Logger() *logrus.Logger {
	if log == nil {
		InitializeSAM3Logger()
	}
	return log
}

func init() {
	InitializeSAM3Logger()
}
=======
import "github.com/go-i2p/go-sam-go/logger"

var log = logger.GetSAM3Logger()

func init() {
	logger.InitializeSAM3Logger()
	log = logger.GetSAM3Logger()
}
>>>>>>> 11110ed5a2ac9ad702c0278886c0418c8cd95a16
