package main

import (
	"github.com/skanehira/docui/panel"

	"github.com/jroimartin/gocui"
)

func main() {
	gui := panel.New(gocui.Output256)
	defer gui.Close()

	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}
