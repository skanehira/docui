package main

import (
	"os"

	"github.com/skanehira/docui/panel"

	"github.com/jroimartin/gocui"
)

func main() {
	for {
		gui := panel.New(gocui.Output256)
		err := gui.MainLoop()

		switch err {
		case gocui.ErrQuit:
			gui.Close()
			os.Exit(0)
		case panel.AttachFlag:
			gui.Gui.Close()
			gui.Panels[panel.ContainerListPanel].(*panel.ContainerList).Attach()
		}
	}
}
