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
	navigate  *navigate
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
	pages *tview.Pages
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

func (g *Gui) volumePanel() *volumes {
	for _, panel := range g.state.panels.panel {
		if panel.name() == "volumes" {
			return panel.(*volumes)
		}
	}
	return nil
}

func (g *Gui) networkPanel() *networks {
	for _, panel := range g.state.panels.panel {
		if panel.name() == "networks" {
			return panel.(*networks)
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

func (g *Gui) monitoringTask() {
	common.Logger.Info("start monitoring task")
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
		case <-g.state.stopChans["task"]:
			common.Logger.Info("stop monitoring task")
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
	navi := newNavigate()

	g.state.panels.panel = append(g.state.panels.panel, tasks)
	g.state.panels.panel = append(g.state.panels.panel, images)
	g.state.panels.panel = append(g.state.panels.panel, containers)
	g.state.panels.panel = append(g.state.panels.panel, volumes)
	g.state.panels.panel = append(g.state.panels.panel, networks)
	g.state.navigate = navi

	grid := tview.NewGrid().SetRows(2, 0, 0, 0, 0, 0, 2).
		AddItem(info, 0, 0, 1, 1, 0, 0, true).
		AddItem(tasks, 1, 0, 1, 1, 0, 0, true).
		AddItem(images, 2, 0, 1, 1, 0, 0, true).
		AddItem(containers, 3, 0, 1, 1, 0, 0, true).
		AddItem(volumes, 4, 0, 1, 1, 0, 0, true).
		AddItem(networks, 5, 0, 1, 1, 0, 0, true).
		AddItem(navi, 6, 0, 1, 1, 0, 0, true)

	g.pages = tview.NewPages().
		AddAndSwitchToPage("main", grid, true)

	g.app.SetRoot(g.pages, true)
	g.switchPanel("images")
}

func (g *Gui) startMonitoring() {
	stop := make(chan int, 1)
	g.state.stopChans["task"] = stop
	g.state.stopChans["image"] = stop
	g.state.stopChans["volume"] = stop
	g.state.stopChans["network"] = stop
	g.state.stopChans["container"] = stop
	go g.monitoringTask()
	go g.imagePanel().monitoringImages(g)
	go g.networkPanel().monitoringNetworks(g)
	go g.volumePanel().monitoringVolumes(g)
	go g.containerPanel().monitoringContainers(g)
}

func (g *Gui) stopMonitoring() {
	g.state.stopChans["task"] <- 1
	g.state.stopChans["image"] <- 1
	g.state.stopChans["volume"] <- 1
	g.state.stopChans["network"] <- 1
	g.state.stopChans["container"] <- 1
}

// Start start application
func (g *Gui) Start() error {
	g.initPanels()
	g.startMonitoring()
	if err := g.app.Run(); err != nil {
		g.app.Stop()
		return err
	}

	return nil
}

// Stop stop application
func (g *Gui) Stop() error {
	g.stopMonitoring()
	g.app.Stop()
	return nil
}

func (g *Gui) selectedImage() *image {
	row, _ := g.imagePanel().GetSelection()
	if len(g.state.resources.images) == 0 {
		return nil
	}
	if row-1 < 0 {
		return nil
	}

	return g.state.resources.images[row-1]
}

func (g *Gui) selectedContainer() *container {
	row, _ := g.containerPanel().GetSelection()
	if len(g.state.resources.containers) == 0 {
		return nil
	}
	if row-1 < 0 {
		return nil
	}

	return g.state.resources.containers[row-1]
}

func (g *Gui) selectedVolume() *volume {
	row, _ := g.volumePanel().GetSelection()
	if len(g.state.resources.volumes) == 0 {
		return nil
	}
	if row-1 < 0 {
		return nil
	}

	return g.state.resources.volumes[row-1]
}

func (g *Gui) selectedNetwork() *network {
	row, _ := g.networkPanel().GetSelection()
	if len(g.state.resources.networks) == 0 {
		return nil
	}
	if row-1 < 0 {
		return nil
	}

	return g.state.resources.networks[row-1]
}

func (g *Gui) message(message, doneLabel, page string, doneFunc func()) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{doneLabel}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			g.closeAndSwitchPanel("modal", page)
			if buttonLabel == doneLabel {
				doneFunc()
			}
		})

	g.pages.AddAndSwitchToPage("modal", g.modal(modal, 80, 29), true).ShowPage("main")
}

func (g *Gui) confirm(message, doneLabel, page string, doneFunc func()) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{doneLabel, "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			g.closeAndSwitchPanel("modal", page)
			if buttonLabel == doneLabel {
				doneFunc()
			}
		})

	g.pages.AddAndSwitchToPage("modal", g.modal(modal, 80, 29), true).ShowPage("main")
}

func (g *Gui) switchPanel(panelName string) {
	for i, panel := range g.state.panels.panel {
		if panel.name() == panelName {
			g.state.navigate.update(panelName)
			panel.focus(g)
			g.state.panels.currentPanel = i
		} else {
			panel.unfocus()
		}
	}
}

func (g *Gui) closeAndSwitchPanel(removePanel, switchPanel string) {
	g.pages.RemovePage(removePanel).ShowPage("main")
	g.switchPanel(switchPanel)
}

func (g *Gui) modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

func (g *Gui) currentPanel() panel {
	return g.state.panels.panel[g.state.panels.currentPanel]
}
