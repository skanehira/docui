package main

import (
	"flag"
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
	logLevel = flag.String("log", "info", "log level")
)

func run() int {
	docker.NewDocker(docker.NewClientConfig(*endpoint, *cert, *key, *ca, *api))
	common.NewLogger(*logLevel)
	gui := gui.New()

	if err := gui.Start(); err != nil {
		return 2
	}

	return 0
}

func main() {
	os.Exit(run())
}
