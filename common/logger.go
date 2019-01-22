package common

import (
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

type Logger struct {
	*os.File
	*log.Logger
}

func NewLogger() *Logger {
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
	log.SetReportCaller(true)

	return &Logger{
		File:   logFile,
		Logger: log.StandardLogger(),
	}
}

func (l *Logger) CloseLogger() {
	if err := l.Close(); err != nil {
		panic(err)
	}
}
