package gui

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type task struct {
	ID      string
	Name    string
	Status  string
	Created string
	Func    func() error
}

type tasks struct {
	*tview.Table
}

func newTasks(g *Gui) *tasks {
	tasks := &tasks{
		Table: tview.NewTable().SetSelectable(true, false),
	}

	tasks.SetTitle("tasks").SetTitleAlign(tview.AlignLeft)
	tasks.SetBorder(true)
	tasks.setEntries(g)
	return tasks
}

func (c *tasks) name() string {
	return "tasks"
}

func (c *tasks) setKeybinding(f func(event *tcell.EventKey) *tcell.EventKey) {
	c.SetInputCapture(f)
}

func (c *tasks) entries(g *Gui) {
	// do nothing
}

func (c *tasks) setEntries(g *Gui) {
	c.entries(g)
	table := c.Clear().Select(0, 0).SetFixed(1, 1)

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

func (c *tasks) focus(g *Gui) {
	c.SetSelectable(true, false)
	g.app.SetFocus(c)
}

func (c *tasks) unfocus() {
	c.SetSelectable(false, false)
}
