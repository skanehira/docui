package main

import (
	"flag"
	"log"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
	"github.com/skanehira/docui/panel"
)

var (
	endpoint = flag.String("endpoint", "unix:///var/run/docker.sock", "Docker endpoint")
	cert     = flag.String("cert", "", "cert.pem file path")
	key      = flag.String("key", "", "key.pem file path")
	ca       = flag.String("ca", "", "ca.pem file path")
)

func main() {
	// if terminal window size is not zero
	if !common.IsTerminalWindowSizeThanZero() {
		return
	}

	// parse flag
	flag.Parse()

	// new dcoker client
	config := docker.NewClientConfig(*endpoint, *cert, *key, *ca)
	dockerClient := docker.NewDocker(config)

	// when docker client cannot connect engine exit
	if err := dockerClient.Ping(); err != nil {
		log.Println(err)
		return
	}

LOOP:
	for {
		// create new panel
		gui := panel.New(gocui.Output256, dockerClient)
		gui.Logger.Info("docui start")

		// run docui
		err := gui.MainLoop()

		switch err {
		case gocui.ErrQuit:
			// exit
			gui.Logger.Info("docui finished")
			gui.Close()
			break LOOP
		case panel.ErrExecFlag:
			// when exec container gui will return ExecFlag
			gui.Close()
			gui.Panels[panel.ContainerListPanel].(*panel.ContainerList).Exec()
		}
	}
}
