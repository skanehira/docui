package gui

import "github.com/gdamore/tcell"

type panel interface {
	name() string
	entries(*Gui)
	setEntries(*Gui)
	setKeybinding(func(event *tcell.EventKey) *tcell.EventKey)
}
