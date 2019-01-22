package panel

import (
	"strconv"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

type TaskStatus string

var (
	Success   TaskStatus = "Success"
	Executing TaskStatus = "Executing"
	Error     TaskStatus = "Error"
)

func (t TaskStatus) String() string {
	return string(t)
}

type TaskList struct {
	*Gui
	name string
	Position
	Tasks    chan *Task
	ViewTask []*Task
	view     *gocui.View
}

type Task struct {
	ID      string
	Task    string `tag:"TASK" len:"min:0.3 max:0.3"`
	Status  string `tag:"STATUS" len:"min:0.3 max:0.3"`
	Created string `tag:"CREATED" len:"min:0.3 max:0.3"`
	Func    func() error
}

func NewTask(task string, function func() error) *Task {
	return &Task{
		Task:    task,
		Status:  Executing.String(),
		Created: common.DateNow(),
		Func:    function,
	}
}

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
	}
}

func (t *TaskList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := g.SetView(TaskListHeaderPanel, t.x, t.y, t.w, t.h); err != nil {
		if err != gocui.ErrUnknownView {
			t.Logger.Error(err)
			return err
		}

		v.Wrap = true
		v.Frame = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormatedHeader(v, &Task{})
	}

	// set scroll panel
	v, err := g.SetView(t.name, t.x, t.y+1, t.w, t.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			t.Logger.Error(err)
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

	go t.MonitorTaskList(t.Tasks)

	return nil
}

func (t *TaskList) Name() string {
	return t.name
}

func (t *TaskList) Refresh(g *gocui.Gui, v *gocui.View) error {
	// do nothing
	return nil
}

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

func (t *TaskList) SetKeyBinding() {
	t.SetKeyBindingToPanel(t.name)
	// TODO add detail and cancel key mapping
}

func (t *TaskList) MonitorTaskList(task chan *Task) {
	for {
		select {
		case task := <-task:
			if err := task.Func(); err != nil {
				task.Status = err.Error()
			} else {
				task.Status = Success.String()
			}

			t.UpdateTask(task)
		}
	}
}

func (t *TaskList) StartTask(task *Task) {
	task.ID = strconv.Itoa(len(t.ViewTask))

	t.ViewTask = append(t.ViewTask, task)

	// status update executing
	t.UpdateTask(task)

	t.Tasks <- task
}

func (t *TaskList) UpdateTask(task *Task) {
	for _, vtask := range t.ViewTask {
		if vtask.ID == task.ID {
			vtask.Status = task.Status
		}
	}

	t.Update(func(g *gocui.Gui) error {
		t.view.Clear()
		for _, task := range t.ViewTask {
			common.OutputFormatedLine(t.view, task)
		}

		return nil
	})
}

func (t *TaskList) CancelTask(id string) error {
	// TODO cancel task

	return nil
}
