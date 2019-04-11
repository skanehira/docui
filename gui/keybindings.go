package gui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func (g *Gui) setGlobalKeybinding(event *tcell.EventKey) {
	switch event.Rune() {
	case 'h':
		g.prevPanel()
	case 'l':
		g.nextPanel()
	}

	switch event.Key() {
	case tcell.KeyTab:
		g.nextPanel()
	case tcell.KeyBacktab:
		g.prevPanel()
	case tcell.KeyRight:
		g.nextPanel()
	case tcell.KeyLeft:
		g.prevPanel()
	}
}

func (g *Gui) nextPanel() {
	idx := (g.state.panels.currentPanel + 1) % len(g.state.panels.panel)
	g.switchPanel(g.state.panels.panel[idx].name())
}

func (g *Gui) prevPanel() {
	g.state.panels.currentPanel--

	if g.state.panels.currentPanel < 0 {
		g.state.panels.currentPanel = len(g.state.panels.panel) - 1
	}

	idx := (g.state.panels.currentPanel) % len(g.state.panels.panel)
	g.switchPanel(g.state.panels.panel[idx].name())
}

func (g *Gui) switchPanel(panelName string) {
	for i, panel := range g.state.panels.panel {
		if panel.name() == panelName {
			panel.focus(g)
			g.state.panels.currentPanel = i
		} else {
			panel.unfocus()
		}
	}
}

func (g *Gui) createContainer() {
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewGrid().
			SetColumns(0, width, 0).
			SetRows(0, height, 0).
			AddItem(p, 1, 1, 1, 1, 0, 0, true)
	}

	form := tview.NewForm()
	form.SetBorder(true)
	form.AddInputField("Name", "", 70, nil, nil).
		AddInputField("HostIP", "", 70, nil, nil).
		AddInputField("Port", "", 70, nil, nil).
		AddDropDown("VolumeType", []string{"bind", "volume"}, 0, func(option string, optionIndex int) {}).
		AddInputField("HostVolume", "", 70, nil, nil).
		AddInputField("Volume", "", 70, nil, nil).
		AddInputField("Image", "", 70, nil, nil).
		AddInputField("User", "", 70, nil, nil).
		AddCheckbox("Attach", false, nil).
		AddButton("Save", func() {
			// TODO get input value, create a container and close
			// form.GetFormItemByLabel("Name").(*tview.InputField).GetText()

			g.pages.RemovePage("form")
			g.switchPanel("images")
		}).
		AddButton("Quit", func() {
			g.pages.RemovePage("form")
			g.switchPanel("images")
		})

	g.pages.AddAndSwitchToPage("form", modal(form, 80, 23), true)
	g.pages.ShowPage("main")
}
