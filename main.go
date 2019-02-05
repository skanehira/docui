package main

import (
	"flag"
	"os"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/docker"
	"github.com/skanehira/docui/panel"
)

func main() {
	var (
		endpoint = flag.String("endpoint", "unix:///var/run/docker.sock", "Docker endpoint")
		cert     = flag.String("cert", "", "cert.pem file path")
		key      = flag.String("key", "", "key.pem file path")
		ca       = flag.String("ca", "", "ca.pem file path")
	)
	config := docker.NewClientConfig(*endpoint, *cert, *key, *ca)
	dockerClient := docker.NewDocker(config)

	for {
		gui := panel.New(gocui.Output256, dockerClient)
		gui.Logger.Info("docui start")
		err := gui.MainLoop()

		switch err {
		case gocui.ErrQuit:
			gui.Logger.Info("docui finished")
			gui.Close()
			os.Exit(0)
		case panel.ExecFlag:
			gui.Gui.Close()
			gui.Panels[panel.ContainerListPanel].(*panel.ContainerList).Exec()
		}
	}
}
