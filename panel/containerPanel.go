package panel

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

// ContainerList container list panel.
type ContainerList struct {
	*Gui
	name string
	Position
	Containers []*Container
	Data       map[string]interface{}
	filter     string
	form       *Form
	stop       chan int
}

// Container container info.
type Container struct {
	ID      string `tag:"ID" len:"min:0.1 max:0.2"`
	Name    string `tag:"NAME" len:"min:0.1 max:0.2"`
	Image   string `tag:"IMAGE" len:"min:0.1 max:0.2"`
	Status  string `tag:"STATUS" len:"min:0.1 max:0.1"`
	Created string `tag:"CREATED" len:"min:0.1 max:0.1"`
	Port    string `tag:"PORT" len:"min:0.1 max:0.2"`
}

// NewContainerList create new container list panel.
func NewContainerList(gui *Gui, name string, x, y, w, h int) *ContainerList {
	return &ContainerList{
		Gui:      gui,
		name:     name,
		Position: Position{x, y, w, h},
		Data:     make(map[string]interface{}),
		stop:     make(chan int, 1),
	}
}

// Name get panel name.
func (c *ContainerList) Name() string {
	return c.name
}

// Edit filtering container list.
func (c *ContainerList) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
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

	c.filter = ReadViewBuffer(v)

	if v, err := c.View(c.name); err == nil {
		c.GetContainerList(v)
	}
}

// SetView set up container list panel.
func (c *ContainerList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := common.SetViewWithValidPanelSize(g, ContainerListHeaderPanel, c.x, c.y, c.w, c.h); err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}

		v.Wrap = true
		v.Frame = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormattedHeader(v, &Container{})
	}

	// set scroll panel
	v, err := common.SetViewWithValidPanelSize(g, c.name, c.x, c.y+1, c.w, c.h)
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

		c.GetContainerList(v)
	}

	c.SetKeyBinding()

	// monitoring container status.
	go c.Monitoring(c.stop, c.Gui.Gui, v)
	return nil
}

// Monitoring monitoring image list.
func (c *ContainerList) Monitoring(stop chan int, g *gocui.Gui, v *gocui.View) {
	common.Logger.Info("monitoring container list start")
	ticker := time.NewTicker(5 * time.Second)

LOOP:
	for {
		select {
		case <-ticker.C:
			c.Update(func(g *gocui.Gui) error {
				return c.Refresh(g, v)
			})
		case <-stop:
			ticker.Stop()
			break LOOP
		}
	}
	common.Logger.Info("monitoring container list stop")
}

// CloseView close panel
func (c *ContainerList) CloseView() {
	// stop monitoring
	c.stop <- 0
	close(c.stop)
}

// SetKeyBinding set key bind to this panel.
func (c *ContainerList) SetKeyBinding() {
	c.SetKeyBindingToPanel(c.name)

	if err := c.SetKeybinding(c.name, gocui.KeyEnter, gocui.ModNone, c.DetailContainer); err != nil {
		panic(err)
	}
	if err := c.SetKeybinding(c.name, 'o', gocui.ModNone, c.DetailContainer); err != nil {
		panic(err)
	}
	if err := c.SetKeybinding(c.name, 'd', gocui.ModNone, c.RemoveContainer); err != nil {
		panic(err)
	}
	if err := c.SetKeybinding(c.name, 'u', gocui.ModNone, c.StartContainer); err != nil {
		panic(err)
	}
	if err := c.SetKeybinding(c.name, 's', gocui.ModNone, c.StopContainer); err != nil {
		panic(err)
	}
	if err := c.SetKeybinding(c.name, 'e', gocui.ModNone, c.ExportContainerPanel); err != nil {
		panic(err)
	}
	if err := c.SetKeybinding(c.name, 'c', gocui.ModNone, c.CommitContainerPanel); err != nil {
		panic(err)
	}
	if err := c.SetKeybinding(c.name, 'r', gocui.ModNone, c.RenameContainerPanel); err != nil {
		panic(err)
	}
	if err := c.SetKeybinding(c.name, gocui.KeyCtrlR, gocui.ModNone, c.Refresh); err != nil {
		panic(err)
	}
	if err := c.SetKeybinding(c.name, 'f', gocui.ModNone, c.Filter); err != nil {
		panic(err)
	}
	if err := c.SetKeybinding(c.name, gocui.KeyCtrlC, gocui.ModNone, c.ExecContainerCmd); err != nil {
		panic(err)
	}
}

// selected return selected container info
func (c *ContainerList) selected() (*Container, error) {
	v, _ := c.View(c.name)
	_, cy := v.Cursor()
	_, oy := v.Origin()

	index := oy + cy
	length := len(c.Containers)

	if index >= length {
		return nil, common.ErrNoContainer
	}
	return c.Containers[cy+oy], nil
}

// DetailContainer display the container detail info
func (c *ContainerList) DetailContainer(g *gocui.Gui, v *gocui.View) error {
	common.Logger.Info("inspect container start")
	defer common.Logger.Info("inspect container end")

	selected, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		common.Logger.Error(err)
		return nil
	}

	container, err := c.Docker.InspectContainer(selected.ID)
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		common.Logger.Error(err)
		return nil
	}

	c.PopupDetailPanel(g, v)

	v, err = g.View(DetailPanel)
	if err != nil {
		common.Logger.Error(err)
		return nil
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)
	fmt.Fprint(v, common.StructToJSON(container))

	return nil
}

// RemoveContainer remove the specified container.
func (c *ContainerList) RemoveContainer(g *gocui.Gui, v *gocui.View) error {
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		common.Logger.Error(err)
		return nil
	}

	c.ConfirmMessage("Are you sure you want to remove this container?", c.name, func() error {
		c.AddTask(fmt.Sprintf("Remove container %s", container.Name), func() error {
			common.Logger.Info("remove container start")
			defer common.Logger.Info("remove container end")

			if err := c.Docker.RemoveContainer(container.ID); err != nil {
				c.ErrMessage(err.Error(), c.name)
				common.Logger.Error(err)
				return err
			}

			return c.Refresh(g, v)
		})
		return nil
	})

	return nil
}

// StartContainer start the specified container.
func (c *ContainerList) StartContainer(g *gocui.Gui, v *gocui.View) error {
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		common.Logger.Error(err)
		return nil
	}

	c.AddTask(fmt.Sprintf("Start container %s", container.Name), func() error {
		common.Logger.Info("start container start")
		defer common.Logger.Info("start container end")

		if err := c.Docker.StartContainer(container.ID); err != nil {
			common.Logger.Error(err)
			return err
		}
		return c.Refresh(g, v)
	})

	return nil
}

// StopContainer stop the specified container.
func (c *ContainerList) StopContainer(g *gocui.Gui, v *gocui.View) error {
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		common.Logger.Error(err)
		return nil
	}

	c.AddTask(fmt.Sprintf("Stop container %s", container.Name), func() error {
		common.Logger.Info("stop container start")
		defer common.Logger.Info("stop container end")

		if err := c.Docker.StopContainer(container.ID); err != nil {
			common.Logger.Error(err)
			return err
		}
		return c.Refresh(g, v)
	})

	return nil
}

// ExportContainerPanel display export container form.
func (c *ContainerList) ExportContainerPanel(g *gocui.Gui, v *gocui.View) error {
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		common.Logger.Error(err)
		return nil
	}

	// get position
	maxX, maxY := c.Size()
	x := maxX / 6
	y := maxY / 3
	w := x * 4

	labelw := 11
	fieldw := w - labelw

	// new form
	form := NewForm(g, ExportContainerPanel, x, y, w, 0)
	c.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		c.SwitchPanel(c.name)
		return nil
	})

	// add fields
	form.AddInput("Path", labelw, fieldw).
		AddValidate(Require.Message+"Path", Require.Validate)
	form.AddInput("Container", labelw, fieldw).
		SetText(container.Name).
		AddValidate(Require.Message+"Container", Require.Validate)
	form.AddButton("OK", c.ExportContainer)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()

	return nil
}

// ExportContainer export specified container
func (c *ContainerList) ExportContainer(g *gocui.Gui, v *gocui.View) error {
	if !c.form.Validate() {
		return nil
	}

	data := c.form.GetFieldTexts()
	container := data["Container"]
	path := data["Path"]

	c.form.Close(g, v)

	c.AddTask(fmt.Sprintf("Export container %s to %s", container, path), func() error {
		common.Logger.Info("export container start")
		defer common.Logger.Info("export container end")

		return c.Docker.ExportContainer(container, path)
	})

	return nil
}

// CommitContainerPanel display commit container form.
func (c *ContainerList) CommitContainerPanel(g *gocui.Gui, v *gocui.View) error {
	// get selected container
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		common.Logger.Error(err)
		return nil
	}

	// get position
	maxX, maxY := c.Size()
	x := maxX / 6
	y := maxY / 3
	w := x * 4

	labelw := 11
	fieldw := w - labelw

	// new form
	form := NewForm(g, CommitContainerPanel, x, y, w, 0)
	c.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		c.SwitchPanel(c.name)
		return nil
	})

	// add fields
	form.AddInput("Repository", labelw, fieldw).
		AddValidate(Require.Message+"Repository", Require.Validate)
	form.AddInput("Tag", labelw, fieldw)
	form.AddInput("Container", labelw, fieldw).
		SetText(container.Name).
		AddValidate(Require.Message+"Container", Require.Validate)
	form.AddButton("OK", c.CommitContainer)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()
	return nil
}

// CommitContainer commit the specified container.
func (c *ContainerList) CommitContainer(g *gocui.Gui, v *gocui.View) error {
	if !c.form.Validate() {
		return nil
	}
	data := c.form.GetFieldTexts()

	if data["Tag"] == "" {
		data["Tag"] = "latest"
	}

	container := data["Container"]
	repository := data["Repository"]
	tag := data["Tag"]

	c.form.Close(g, v)

	c.AddTask(fmt.Sprintf("Commit container %s to %s", container, repository+":"+tag), func() error {
		common.Logger.Info("commit container start")
		defer common.Logger.Info("commit container end")

		if err := c.Docker.CommitContainer(container, types.ContainerCommitOptions{Reference: repository + ":" + tag}); err != nil {
			common.Logger.Error(err)
			return err
		}
		return c.Refresh(g, v)
	})

	return nil
}

// RenameContainerPanel display rename container form.
func (c *ContainerList) RenameContainerPanel(g *gocui.Gui, v *gocui.View) error {
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		common.Logger.Error(err)
		return nil
	}

	//get position
	maxX, maxY := c.Size()
	x := maxX / 8
	y := maxY / 3
	w := x * 6

	labelw := 11
	fieldw := w - labelw

	// new form
	form := NewForm(g, RenameContainerPanel, x, y, w, 0)
	c.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		c.SwitchPanel(c.name)
		return nil
	})

	// add fields
	form.AddInput("NewName", labelw, fieldw).
		AddValidate(Require.Message+"NewName", Require.Validate)
	form.AddInput("Container", labelw, fieldw).
		SetText(container.Name).
		AddValidate(Require.Message+"Container", Require.Validate)
	form.AddButton("OK", c.RenameContainer)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()
	return nil
}

// RenameContainer rename the specified container.
func (c *ContainerList) RenameContainer(g *gocui.Gui, v *gocui.View) error {
	if !c.form.Validate() {
		return nil
	}

	data := c.form.GetFieldTexts()
	oldName := data["Container"]
	newName := data["NewName"]

	c.form.Close(g, v)

	c.AddTask(fmt.Sprintf("Rename container %s to %s", oldName, newName), func() error {
		common.Logger.Info("rename container start")
		defer common.Logger.Info("rename container end")

		if err := c.Docker.RenameContainer(oldName, newName); err != nil {
			common.Logger.Error(err)
			return err
		}
		return c.Refresh(g, v)
	})

	return nil
}

// Refresh update containers info
func (c *ContainerList) Refresh(g *gocui.Gui, v *gocui.View) error {
	c.Update(func(g *gocui.Gui) error {
		v, err := c.View(c.name)
		if err != nil {
			common.Logger.Error(err)
			return nil
		}

		c.GetContainerList(v)

		return nil
	})

	return nil
}

// GetContainerList return containers info
func (c *ContainerList) GetContainerList(v *gocui.View) {
	v.Clear()
	c.Containers = make([]*Container, 0)

	containers, err := c.Docker.Containers(types.ContainerListOptions{All: true})

	if err != nil {
		common.Logger.Error(err)
		return
	}
	for _, con := range containers {
		name := con.Names[0][1:]
		if c.filter != "" {
			if strings.Index(strings.ToLower(name), strings.ToLower(c.filter)) == -1 {
				continue
			}
		}

		id := con.ID[:12]
		image := con.Image
		status := con.Status
		created := common.ParseDateToString(con.Created)
		port := common.ParsePortToString(con.Ports)

		container := &Container{
			ID:      id,
			Image:   image,
			Name:    name,
			Status:  status,
			Created: created,
			Port:    port,
		}

		c.Containers = append(c.Containers, container)

		common.OutputFormattedLine(v, container)
	}
}

// Filter display filtering form.
func (c *ContainerList) Filter(g *gocui.Gui, lv *gocui.View) error {
	isReset := false
	closePanel := func(g *gocui.Gui, v *gocui.View) error {
		if isReset {
			c.filter = ""
		} else {
			lv.SetCursor(0, 0)
			c.filter = ReadViewBuffer(v)
		}
		if v, err := c.View(c.name); err == nil {
			c.GetContainerList(v)
		}

		if err := g.DeleteView(v.Name()); err != nil {
			common.Logger.Error(err)
			return nil
		}

		g.DeleteKeybindings(v.Name())
		c.SwitchPanel(c.name)
		return nil
	}

	reset := func(g *gocui.Gui, v *gocui.View) error {
		isReset = true
		return closePanel(g, v)
	}

	if err := c.NewFilterPanel(c, reset, closePanel); err != nil {
		common.Logger.Error(err)
		return nil
	}

	return nil
}

// ExecContainerCmd display exec container cmd form.
func (c *ContainerList) ExecContainerCmd(g *gocui.Gui, v *gocui.View) error {
	selected, err := c.selected()
	if err != nil {
		common.Logger.Error(err)
		return nil
	}

	container, err := c.Docker.InspectContainer(selected.ID)
	if err != nil {
		common.Logger.Error(err)
		return nil
	}

	if !container.State.Running {
		msg := fmt.Sprintf("container %s is not runnig", selected.Name)
		c.ErrMessage(msg, c.name)
		return nil
	}

	// get position
	maxX, maxY := c.Size()
	x := maxX / 6
	y := maxY / 4
	w := x * 4

	labelw := 5
	fieldw := w - labelw

	// new form
	form := NewForm(g, ExecContainerCmd, x, y, w, 0)
	c.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		c.SwitchPanel(c.name)
		return nil
	})

	exec := func(*gocui.Gui, *gocui.View) error {
		if !c.form.Validate() {
			return nil
		}
		return ErrExecFlag
	}

	// add fields
	form.AddInput("Cmd", labelw, fieldw).
		AddValidate("no specified Cmd", func(value string) bool {
			return value != ""
		})

	form.AddButton("Exec", exec)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()
	return nil
}

// Exec exec the specified cmd run on container.
func (c *ContainerList) Exec() error {
	common.Logger.Info("exec container start")
	defer common.Logger.Info("exec container end")

	selected, err := c.selected()
	if err != nil {
		common.Logger.Error(err)
		return nil
	}

	if err := c.Docker.AttachExecContainer(selected.ID, c.form.GetFieldText("Cmd")); err != nil {
		common.Logger.Error(err)
		return err
	}

	return nil
}
