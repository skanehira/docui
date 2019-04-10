package gui

import (
	"context"

	"github.com/rivo/tview"
	"github.com/skanehira/docui/common"
)

type panels struct {
	currentPanel int
	panel        []panel
}

// docker resources
type resources struct {
	images     []*image
	containers []*container
	networks   []*network
	volumes    []*volume
	tasks      []*task
}

type state struct {
	panels    panels
	resources resources
	stopChans map[string]chan int
}

func newState() *state {
	return &state{
		stopChans: make(map[string]chan int),
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

func (g *Gui) taskPanel() *tasks {
	for _, panel := range g.state.panels.panel {
		if panel.name() == "tasks" {
			return panel.(*tasks)
		}
	}
	return nil
}

func (g *Gui) monitoringTask(stop chan int) {

LOOP:
	for {
		select {
		case task := <-g.taskPanel().tasks:
			if err := task.Func(task.Ctx); err != nil {
				task.Status = err.Error()
			} else {
				task.Status = success
			}
			g.updateTask(task)
		case <-stop:
			break LOOP
		}
	}
}

func (g *Gui) startTask(taskName string, f func(ctx context.Context) error) {
	ctx, cancel := context.WithCancel(context.Background())

	task := &task{
		Name:    taskName,
		Status:  executing,
		Created: common.DateNow(),
		Func:    f,
		Ctx:     ctx,
		Cancel:  cancel,
	}

	g.state.resources.tasks = append(g.state.resources.tasks, task)
	g.updateTask(task)
	g.taskPanel().tasks <- task
}

func (g *Gui) cancelTask() {
	taskPanel := g.taskPanel()
	row, _ := taskPanel.GetSelection()

	task := g.state.resources.tasks[row-1]
	if task.Status == executing {
		task.Cancel()
		task.Status = cancel
		g.updateTask(task)
	}
}

func (g *Gui) updateTask(task *task) {
	g.app.QueueUpdateDraw(func() {
		g.taskPanel().setEntries(g)
	})
}

func (g *Gui) initPanels() {
	tasks := newTasks(g)
	images := newImages(g)
	containers := newContainers(g)
	volumes := newVolumes(g)
	networks := newNetworks(g)
	info := newInfo()

	g.state.panels.panel = append(g.state.panels.panel, tasks)
	g.state.panels.panel = append(g.state.panels.panel, images)
	g.state.panels.panel = append(g.state.panels.panel, containers)
	g.state.panels.panel = append(g.state.panels.panel, volumes)
	g.state.panels.panel = append(g.state.panels.panel, networks)

	grid := tview.NewGrid().SetRows(2, 0, 0, 0, 0, 0).
		AddItem(info, 0, 0, 1, 1, 0, 0, true).
		AddItem(tasks, 1, 0, 1, 1, 0, 0, true).
		AddItem(images, 2, 0, 1, 1, 0, 0, true).
		AddItem(containers, 3, 0, 1, 1, 0, 0, true).
		AddItem(volumes, 4, 0, 1, 1, 0, 0, true).
		AddItem(networks, 5, 0, 1, 1, 0, 0, true)

	g.app.SetRoot(grid, true)
	g.switchPanel("images")
}

// Start start application
func (g *Gui) Start() error {
	g.initPanels()
	stop := make(chan int, 1)
	g.state.stopChans["task"] = stop
	go g.monitoringTask(stop)

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
