package main

import (
	"os"

	"github.com/skanehira/docui/panel"

	"github.com/jroimartin/gocui"
)

func main() {
	for {
		gui := panel.New(gocui.Output256)
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
