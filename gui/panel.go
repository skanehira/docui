package gui

type panel interface {
	name() string
	entries(*Gui)
	setEntries(*Gui)
	updateEntries(*Gui)
	setKeybinding(*Gui)
	focus(*Gui)
	unfocus()
}
