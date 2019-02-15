package panel

import (
	"strconv"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

// TaskStatus define new type of task status.
type TaskStatus string

var (
	// Success when task success.
	Success TaskStatus = "Success"
	// Executing when task running.
	Executing TaskStatus = "Executing"
	// Error when task is failed.
	Error TaskStatus = "Error"
)

// String return task name.
func (t TaskStatus) String() string {
	return string(t)
}

// TaskList task panel.
type TaskList struct {
	*Gui
	name string
	Position
	Tasks    chan *Task
	ViewTask []*Task
	view     *gocui.View
	stop     chan int
}

// Task task info
type Task struct {
	ID      string
	Task    string `tag:"TASK" len:"min:0.3 max:0.3"`
	Status  string `tag:"STATUS" len:"min:0.3 max:0.3"`
	Created string `tag:"CREATED" len:"min:0.3 max:0.3"`
	Func    func() error
}

// NewTask create new task info
func NewTask(task string, function func() error) *Task {
	return &Task{
		Task:    task,
		Status:  Executing.String(),
		Created: common.DateNow(),
		Func:    function,
	}
}

// NewTaskList create new task list panel.
func NewTaskList(gui *Gui, name string, x, y, w, h int) *TaskList {
	return &TaskList{
		Gui:  gui,
		name: name,
		Position: Position{
			x: x,
			y: y,
			w: w,
			h: h,
		},
		Tasks: make(chan *Task),
		stop:  make(chan int, 1),
	}
}

// SetView set up task list panel.
func (t *TaskList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := common.SetViewWithValidPanelSize(g, TaskListHeaderPanel, t.x, t.y, t.w, t.h); err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}

		v.Wrap = true
		v.Frame = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormattedHeader(v, &Task{})
	}

	// set scroll panel
	v, err := common.SetViewWithValidPanelSize(g, t.name, t.x, t.y+1, t.w, t.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}
		v.Frame = false
		v.Wrap = true
		v.FgColor = gocui.ColorGreen
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)

		t.view = v
	}

	t.SetKeyBinding()

	go t.MonitorTaskList(t.stop, t.Gui.Gui, v)

	return nil
}

// CloseView close panel
func (t *TaskList) CloseView() {
	// stop monitoring
	t.stop <- 0
	close(t.stop)
}

// Name return panel name.
func (t *TaskList) Name() string {
	return t.name
}

// Refresh do nothing.
func (t *TaskList) Refresh(g *gocui.Gui, v *gocui.View) error {
	// do nothing
	return nil
}

// Edit this is default editor.
func (t *TaskList) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
		return
	case key == gocui.KeyArrowRight:
		v.MoveCursor(+1, 0, false)
		return
	}
}

// SetKeyBinding set key bind to this panel.
func (t *TaskList) SetKeyBinding() {
	t.SetKeyBindingToPanel(t.name)
	// TODO add detail and cancel key mapping
}

// MonitorTaskList monitoring task status.
func (t *TaskList) MonitorTaskList(stop chan int, g *gocui.Gui, v *gocui.View) {
	common.Logger.Info("monitoring task list start")
LOOP:
	for {
		select {
		case task := <-t.Tasks:
			if err := task.Func(); err != nil {
				task.Status = err.Error()
			} else {
				task.Status = Success.String()
			}

			t.UpdateTask(task)
		case <-stop:
			break LOOP
		}
	}
	common.Logger.Info("monitoring tasks list stop")
}

// StartTask run the specified task.
func (t *TaskList) StartTask(task *Task) {
	task.ID = strconv.Itoa(len(t.ViewTask))

	// push front
	t.ViewTask = append([]*Task{task}, t.ViewTask...)

	// status update executing
	t.UpdateTask(task)

	t.Tasks <- task
}

// UpdateTask update the specified task info
func (t *TaskList) UpdateTask(task *Task) {
	for _, vtask := range t.ViewTask {
		if vtask.ID == task.ID {
			vtask.Status = task.Status
		}
	}

	t.Update(func(g *gocui.Gui) error {
		t.view.Clear()
		for _, task := range t.ViewTask {
			common.OutputFormattedLine(t.view, task)
		}

		return nil
	})
}

// CancelTask this feature will add in the future
func (t *TaskList) CancelTask(id string) error {
	// TODO cancel task

	return nil
}
