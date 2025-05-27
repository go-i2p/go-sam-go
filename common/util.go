package common

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func IgnorePortError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "missing port in address") {
		log.Debug("Ignoring 'missing port in address' error")
		err = nil
	}
	return err
}

func SplitHostPort(hostport string) (string, string, error) {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		if IgnorePortError(err) == nil {
			log.WithField("host", hostport).Debug("Using full string as host, port set to 0")
			host = hostport
			port = "0"
		}
	}
	log.WithFields(logrus.Fields{
		"host": host,
		"port": port,
	}).Debug("Split host and port")
	return host, port, nil
}

func ExtractPairString(input, value string) string {
	log.WithFields(logrus.Fields{"input": input, "value": value}).Debug("ExtractPairString called")
	parts := strings.Split(input, " ")
	for _, part := range parts {
		log.WithField("part", part).Debug("Checking part")
		if strings.HasPrefix(part, value) {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				log.WithFields(logrus.Fields{"key": kv[0], "value": kv[1]}).Debug("Pair extracted")
				return kv[1]
			}
		}
	}
	log.WithFields(logrus.Fields{"input": input, "value": value}).Debug("No pair found")
	return ""
}

func ExtractPairInt(input, value string) int {
	rv, err := strconv.Atoi(ExtractPairString(input, value))
	if err != nil {
		log.WithFields(logrus.Fields{"input": input, "value": value}).Debug("No pair found")
		return 0
	}
	log.WithField("result", rv).Debug("Pair extracted and converted to int")
	return rv
}

func ExtractDest(input string) string {
	log.WithField("input", input).Debug("ExtractDest called")
	dest := strings.Split(input, " ")[0]
	log.WithField("dest", dest).Debug("Destination extracted")
	return strings.Split(input, " ")[0]
}

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randGen    = rand.New(randSource)
)

func RandPort() (portNumber string, err error) {
	maxAttempts := 30
	for range maxAttempts {
		p := randGen.Intn(55534) + 10000
		port := strconv.Itoa(p)
		if l, e := net.Listen("tcp", net.JoinHostPort("localhost", port)); e != nil {
			continue
		} else {
			defer l.Close()
			if l, e := net.Listen("udp", net.JoinHostPort("localhost", port)); e != nil {
				continue
			} else {
				defer l.Close()
				return strconv.Itoa(l.Addr().(*net.UDPAddr).Port), nil
			}
		}
	}

	return "", fmt.Errorf("unable to find a pair of available tcp and udp ports in %v attempts", maxAttempts)
}
