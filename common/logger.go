package common

import (
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

// Logger logger
var Logger *logger

// Logger logger.
type logger struct {
	*os.File
	*log.Logger
}

// NewLogger create new logger.
func NewLogger(logLevel string) {
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile(filepath.Join(home, "docui.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		panic(err)
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	log.SetOutput(logFile)
	log.SetLevel(level)
	log.SetReportCaller(true)

	Logger = &logger{
		File:   logFile,
		Logger: log.StandardLogger(),
	}
}
