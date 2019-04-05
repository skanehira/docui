package gui

import (
	"github.com/rivo/tview"
)

type state struct {
	grid   *tview.Grid
	panels map[string]panel
}

func newState() *state {
	return &state{
		panels: make(map[string]panel),
	}
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

func (g *Gui) initPanels() {
	images := newImages()
	g.state.panels["images"] = images

	grid := tview.NewGrid().SetRows(2, 2, 2, 2, 2)
	grid.AddItem(images, 0, 0, 5, 1, 0, 0, true)

	g.state.grid = grid
}

func (g *Gui) getImages() *images {
	return g.state.panels["images"].(*images)
}

// Start start application
func (g *Gui) Start() error {
	g.initPanels()

	if err := g.app.SetRoot(g.state.grid, true).SetFocus(g.getImages()).Run(); err != nil {
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
