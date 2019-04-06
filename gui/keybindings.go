package gui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type keybinding struct {
	key interface{}
	f   func()
}

func (g *Gui) addKeybinding(p panel, key interface{}, f func()) {
	keybindings, ok := g.state.keybindings[p]
	if !ok {
		g.state.keybindings[p] = []keybinding{{key, f}}
	} else {
		keybindings = append(keybindings, keybinding{key, f})
		g.state.keybindings[p] = keybindings
	}
}

func (g *Gui) addKeybindings() {
	g.addKeybinding(g.imagePanel(), 'l', func() { g.nextPanel() })
	g.addKeybinding(g.imagePanel(), 'h', func() { g.prevPanel() })
	g.addKeybinding(g.imagePanel(), tcell.KeyTab, func() { g.nextPanel() })
	g.addKeybinding(g.imagePanel(), tcell.KeyBacktab, func() { g.prevPanel() })
	g.addKeybinding(g.imagePanel(), tcell.KeyLeft, func() { g.prevPanel() })
	g.addKeybinding(g.imagePanel(), tcell.KeyRight, func() { g.nextPanel() })

	g.addKeybinding(g.containerPanel(), 'l', func() { g.nextPanel() })
	g.addKeybinding(g.containerPanel(), 'h', func() { g.prevPanel() })
	g.addKeybinding(g.containerPanel(), tcell.KeyTab, func() { g.nextPanel() })
	g.addKeybinding(g.containerPanel(), tcell.KeyBacktab, func() { g.prevPanel() })
	g.addKeybinding(g.containerPanel(), tcell.KeyLeft, func() { g.prevPanel() })
	g.addKeybinding(g.containerPanel(), tcell.KeyRight, func() { g.nextPanel() })
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
