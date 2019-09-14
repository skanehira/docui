package common

import (
	"io"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

// Logger logger
var Logger *logger

// Logger logger.
type logger struct {
	io.Writer
	*log.Logger
}

// NewLogger create new logger.
func NewLogger(logLevel string, logFileName string) {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}

	var logWriter io.Writer

	if logFileName == "" {
		logWriter = ioutil.Discard
	} else {
		logWriter, err = os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetLevel(level)
	log.SetOutput(logWriter)
	log.SetReportCaller(true)

	Logger = &logger{
		Writer: logWriter,
		Logger: log.StandardLogger(),
	}
}
