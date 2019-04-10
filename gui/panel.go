package gui

type panel interface {
	name() string
	entries(*Gui)
	setEntries(*Gui)
	setKeybinding(*Gui)
	focus(*Gui)
	unfocus()
}
