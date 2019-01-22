package common

import (
	"os"
	"path/filepath"

	"os/user"

	log "github.com/sirupsen/logrus"
)

type Logger struct {
	*os.File
	*log.Logger
}

func NewLogger() *Logger {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	logFile, err := os.OpenFile(filepath.Join(user.HomeDir, "docui.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

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
