package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
	"github.com/skanehira/docui/gui"
)

var (
	endpoint = flag.String("endpoint", "unix:///var/run/docker.sock", "Docker endpoint")
	cert     = flag.String("cert", "", "cert.pem file path")
	key      = flag.String("key", "", "key.pem file path")
	ca       = flag.String("ca", "", "ca.pem file path")
	api      = flag.String("api", "1.39", "api version")
	logFile  = flag.String("log", "", "log file path")
	logLevel = flag.String("log-level", "info", "log level")
)

func run() int {
	common.NewLogger(*logLevel, *logFile)

	docker.NewDocker(docker.NewClientConfig(*endpoint, *cert, *key, *ca, *api))
	if _, err := docker.Client.Info(context.TODO()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	gui := gui.New()

	if err := gui.Start(); err != nil {
		common.Logger.Errorf("cannot start docui: %s", err)
		return 1
	}

	return 0
}

func main() {
	flag.Parse()
	os.Exit(run())
}
