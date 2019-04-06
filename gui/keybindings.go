package gui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func (g *Gui) addKeybinding(p panel, key interface{}, f func()) {
	keybindings, ok := g.state.keybindings[p]
	if !ok {
		g.state.keybindings[p] = []keybinding{{key, f}}
	} else {
		keybindings = append(keybindings, keybinding{key, f})
		g.state.keybindings[p] = keybindings
	}
}

func (g *Gui) addGlobalKeybindings() {
	keybindings := []struct {
		key interface{}
		f   func()
	}{
		{'l', func() { g.nextPanel() }},
		{'h', func() { g.prevPanel() }},
		{tcell.KeyTab, func() { g.nextPanel() }},
		{tcell.KeyBacktab, func() { g.prevPanel() }},
		{tcell.KeyLeft, func() { g.prevPanel() }},
		{tcell.KeyRight, func() { g.nextPanel() }},
	}

	for _, keybind := range keybindings {
		for _, panel := range g.state.panels.panel {
			g.addKeybinding(panel, keybind.key, keybind.f)
		}
	}
}

func (g *Gui) addKeybindings() {
	g.addGlobalKeybindings()
}

func (g *Gui) setKeybindings() {
	g.addKeybindings()

	for panel, keybindings := range g.state.keybindings {
		panel.setKeybinding(func(event *tcell.EventKey) *tcell.EventKey {
			for _, keybind := range keybindings {
				key, ok := keybind.key.(tcell.Key)

				if ok {
					if event.Key() == key {
						keybind.f()
					}
				} else {
					if event.Rune() == keybind.key.(rune) {
						keybind.f()
					}
				}
			}
			return event
		})
	}
}

func (g *Gui) nextPanel() {
	idx := (g.state.panels.currentPanel + 1) % len(g.state.panels.panel)
	g.state.panels.currentPanel = idx
	g.app.SetFocus(g.state.panels.panel[idx].(tview.Primitive))
}

func (g *Gui) prevPanel() {
	g.state.panels.currentPanel--

	if g.state.panels.currentPanel < 0 {
		g.state.panels.currentPanel = len(g.state.panels.panel) - 1
	} else {
	}

	idx := (g.state.panels.currentPanel) % len(g.state.panels.panel)
	g.state.panels.currentPanel = idx
	g.app.SetFocus(g.state.panels.panel[idx].(tview.Primitive))
}

func (g *Gui) switchPanel(panelName string) {
	for _, panel := range g.state.panels.panel {
		if panel.name() == panelName {
			g.app.SetFocus(panel.(tview.Primitive))
		}
	}
}
