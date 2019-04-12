package gui

import (
	"context"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var (
	success   = "Success"
	executing = "Executing"
	cancel    = "canceled"
)

type task struct {
	Name    string
	Status  string
	Created string
	Func    func(ctx context.Context) error
	Ctx     context.Context
	Cancel  context.CancelFunc
}

type tasks struct {
	*tview.Table
	tasks chan *task
}

func newTasks(g *Gui) *tasks {
	tasks := &tasks{
		Table: tview.NewTable().SetSelectable(true, false).Select(0, 0).SetFixed(1, 1),
		tasks: make(chan *task),
	}

	tasks.SetTitle("tasks").SetTitleAlign(tview.AlignLeft)
	tasks.SetBorder(true)
	tasks.setEntries(g)
	tasks.setKeybinding(g)
	return tasks
}

func (t *tasks) name() string {
	return "tasks"
}

func (t *tasks) setKeybinding(g *Gui) {
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		g.setGlobalKeybinding(event)

		switch event.Key() {
		}

		switch event.Rune() {
		}

		return event
	})
}

func (t *tasks) entries(g *Gui) {
	// do nothing
}

func (t *tasks) setEntries(g *Gui) {
	t.entries(g)
	table := t.Clear()

	headers := []string{
		"Name",
		"Status",
		"Created",
	}

	for i, header := range headers {
		table.SetCell(0, i, &tview.TableCell{
			Text:            header,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorWhite,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrBold,
		})
	}

	for i, task := range g.state.resources.tasks {
		table.SetCell(i+1, 0, tview.NewTableCell(task.Name).
			SetTextColor(tcell.ColorLightGreen).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 1, tview.NewTableCell(task.Status).
			SetTextColor(tcell.ColorLightGreen).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 2, tview.NewTableCell(task.Created).
			SetTextColor(tcell.ColorLightGreen).
			SetMaxWidth(1).
			SetExpansion(1))

	}
}

func (t *tasks) focus(g *Gui) {
	t.SetSelectable(true, false)
	g.app.SetFocus(t)
}

func (t *tasks) unfocus() {
	t.SetSelectable(false, false)
}

func (t *tasks) updateEntries(g *Gui) {
	// do nothings
}
