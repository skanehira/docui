package main

import (
	"flag"
	"os"

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
	docker.Client = docker.NewDocker(docker.NewClientConfig(*endpoint, *cert, *key, *ca, *api))
	gui := gui.New()
	if err := gui.Start(); err != nil {
		return 2
	}

	return 0
}

func main() {
	os.Exit(run())
}

//func main() {
//	// if terminal window size is not zero
//	if !common.IsTerminalWindowSizeThanZero() {
//		return
//	}
//
//	// create logger
//	common.NewLogger(*logLevel)
//
//	// parse flag
//	flag.Parse()
//
//	// new docker client
//	config := docker.NewClientConfig(*endpoint, *cert, *key, *ca, *api)
//	dockerClient := docker.NewDocker(config)
//
//	// when docker client cannot connect engine exit
//	info, err := dockerClient.Ping(context.TODO())
//	if err != nil {
//		common.Logger.Error(err)
//		panic(err)
//	}
//
//	common.Logger.Infof("docker engine info: %#+v", info)
//
//LOOP:
//	for {
//		// create new panel
//		gui := panel.New(gocui.Output256, dockerClient)
//		common.Logger.Info("docui start")
//
//		// run docui
//		err := gui.MainLoop()
//
//		switch err {
//		case gocui.ErrQuit:
//			// exit
//			common.Logger.Info("docui end")
//			gui.Close()
//			break LOOP
//		case panel.ErrExecFlag:
//			// when exec container gui will return ExecFlag
//			gui.Close()
//			gui.Panels[panel.ContainerListPanel].(*panel.ContainerList).Exec()
//		case panel.ErrLogsFlag:
//			// when show container logs gui will return ExecFlag
//			gui.Close()
//			gui.Panels[panel.ContainerListPanel].(*panel.ContainerList).ShowContainerLogs()
//		}
//	}
//}
