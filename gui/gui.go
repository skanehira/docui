package gui

import (
	"github.com/rivo/tview"
)

type panels struct {
	currentPanel int
	panel        []panel
}

type resources struct {
	images     []*image
	containers []*container
	networks   []*network
	volumes    []*volume
}

type keybinding struct {
	key interface{}
	f   func()
}

type state struct {
	panels      panels
	resources   resources
	keybindings map[panel][]keybinding
}

func newState() *state {
	return &state{keybindings: make(map[panel][]keybinding)}
}

// Gui have all panels
type Gui struct {
	app   *tview.Application
	state *state
}

// New create new gui
func New() *Gui {
	return &Gui{
		app:   tview.NewApplication(),
		state: newState(),
	}
}

func (g *Gui) imagePanel() *images {
	for _, panel := range g.state.panels.panel {
		if panel.name() == "images" {
			return panel.(*images)
		}
	}
	return nil
}

func (g *Gui) containerPanel() *containers {
	for _, panel := range g.state.panels.panel {
		if panel.name() == "containers" {
			return panel.(*containers)
		}
	}
	return nil
}

func (g *Gui) initPanels() {
	images := newImages(g)
	containers := newContainers(g)
	volumes := newVolumes(g)
	networks := newNetworks(g)
	info := newInfo()

	g.state.panels.panel = append(g.state.panels.panel, images)
	g.state.panels.panel = append(g.state.panels.panel, containers)
	g.state.panels.panel = append(g.state.panels.panel, volumes)
	g.state.panels.panel = append(g.state.panels.panel, networks)

	grid := tview.NewGrid().SetRows(2, 0, 0, 0, 0).
		AddItem(info, 0, 0, 1, 1, 0, 0, true).
		AddItem(images, 1, 0, 1, 1, 0, 0, true).
		AddItem(containers, 2, 0, 1, 1, 0, 0, true).
		AddItem(volumes, 3, 0, 1, 1, 0, 0, true).
		AddItem(networks, 4, 0, 1, 1, 0, 0, true)

	g.app.SetRoot(grid, true)
	g.switchPanel("images")
}

// Start start application
func (g *Gui) Start() error {
	g.initPanels()
	g.setKeybindings()

	if err := g.app.Run(); err != nil {
		g.app.Stop()
		return err
	}

	return nil
}

// Stop stop application
func (g *Gui) Stop() error {
	g.app.Stop()
	return nil
}
