package panel

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
	"golang.org/x/crypto/ssh/terminal"
)

type ContainerList struct {
	*Gui
	name string
	Position
	Containers []*Container
	Data       map[string]interface{}
	filter     string
	form       *Form
}

type Container struct {
	ID      string `tag:"ID" len:"min:0.1 max:0.2"`
	Name    string `tag:"NAME" len:"min:0.1 max:0.2"`
	Image   string `tag:"IMAGE" len:"min:0.1 max:0.2"`
	Status  string `tag:"STATUS" len:"min:0.1 max:0.1"`
	Created string `tag:"CREATED" len:"min:0.1 max:0.1"`
	Port    string `tag:"PORT" len:"min:0.1 max:0.2"`
}

func NewContainerList(gui *Gui, name string, x, y, w, h int) *ContainerList {
	return &ContainerList{
		Gui:      gui,
		name:     name,
		Position: Position{x, y, w, h},
		Data:     make(map[string]interface{}),
	}
}

func (c *ContainerList) Name() string {
	return c.name
}

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

func (c *ContainerList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := g.SetView(ContainerListHeaderPanel, c.x, c.y, c.w, c.h); err != nil {
		if err != gocui.ErrUnknownView {
			c.Logger.Error(err)
			return err
		}

		v.Wrap = true
		v.Frame = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormatedHeader(v, &Container{})
	}

	// set scroll panel
	v, err := g.SetView(c.name, c.x, c.y+1, c.w, c.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			c.Logger.Error(err)
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

	//monitoring container status interval 5s
	go func() {
		for {
			c.Update(func(g *gocui.Gui) error {
				c.Refresh(g, v)
				return nil
			})
			time.Sleep(5 * time.Second)
		}
	}()

	return nil
}

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
	if err := c.SetKeybinding(c.name, 'a', gocui.ModNone, c.AttachContainer); err != nil {
		panic(err)
	}
}

func (c *ContainerList) selected() (*Container, error) {
	v, _ := c.View(c.name)
	_, cy := v.Cursor()
	_, oy := v.Origin()

	index := oy + cy
	length := len(c.Containers)

	if index >= length {
		return nil, common.NoContainer
	}
	return c.Containers[cy+oy], nil
}

func (c *ContainerList) DetailContainer(g *gocui.Gui, v *gocui.View) error {
	c.Logger.Info("inspect container start")
	defer c.Logger.Info("inspect container finished")

	selected, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		c.Logger.Error(err)
		return nil
	}

	container, err := c.Docker.InspectContainer(selected.ID)
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		c.Logger.Error(err)
		return nil
	}

	c.PopupDetailPanel(g, v)

	v, err = g.View(DetailPanel)
	if err != nil {
		c.Logger.Error(err)
		return nil
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)
	fmt.Fprint(v, common.StructToJson(container))

	return nil
}

func (c *ContainerList) RemoveContainer(g *gocui.Gui, v *gocui.View) error {
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		c.Logger.Error(err)
		return nil
	}

	c.ConfirmMessage("Are you sure you want to remove this container?", c.name, func() error {
		c.AddTask(fmt.Sprintf("Remove container %s", container.Name), func() error {
			c.Logger.Info("remove container start")
			defer c.Logger.Info("remove container finished")

			options := docker.RemoveContainerOptions{ID: container.ID}

			if err := c.Docker.RemoveContainer(options); err != nil {
				c.ErrMessage(err.Error(), c.name)
				c.Logger.Error(err)
				return err
			}

			return c.Refresh(g, v)
		})
		return nil
	})

	return nil
}

func (c *ContainerList) StartContainer(g *gocui.Gui, v *gocui.View) error {
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		c.Logger.Error(err)
		return nil
	}

	c.AddTask(fmt.Sprintf("Start container %s", container.Name), func() error {
		c.Logger.Info("start container start")
		defer c.Logger.Info("start container finished")

		if err := c.Docker.StartContainerWithID(container.ID); err != nil {
			c.Logger.Error(err)
			return err
		}
		return c.Refresh(g, v)
	})

	return nil
}

func (c *ContainerList) StopContainer(g *gocui.Gui, v *gocui.View) error {
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		c.Logger.Error(err)
		return nil
	}

	c.AddTask(fmt.Sprintf("Stop container %s", container.Name), func() error {
		c.Logger.Info("stop container start")
		defer c.Logger.Info("stop container finished")

		if err := c.Docker.StopContainerWithID(container.ID); err != nil {
			c.Logger.Error(err)
			return err
		}
		return c.Refresh(g, v)
	})

	return nil
}

func (c *ContainerList) ExportContainerPanel(g *gocui.Gui, v *gocui.View) error {
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		c.Logger.Error(err)
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

func (c *ContainerList) ExportContainer(g *gocui.Gui, v *gocui.View) error {
	if !c.form.Validate() {
		return nil
	}

	data := c.form.GetFieldTexts()
	container := data["Container"]
	path := data["Path"]

	c.form.Close(g, v)

	c.AddTask(fmt.Sprintf("Export container %s to %s", container, path), func() error {
		c.Logger.Info("export container start")
		defer c.Logger.Info("export container finished")

		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			c.Logger.Error(err)
			return err
		}
		defer file.Close()

		options := docker.ExportContainerOptions{
			ID:           container,
			OutputStream: file,
		}
		return c.Docker.ExportContainerWithOptions(options)
	})

	return nil
}

func (c *ContainerList) CommitContainerPanel(g *gocui.Gui, v *gocui.View) error {
	// get selected container
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		c.Logger.Error(err)
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

	options := docker.CommitContainerOptions{
		Container:  container,
		Repository: repository,
		Tag:        tag,
	}

	c.form.Close(g, v)

	c.AddTask(fmt.Sprintf("Commit container %s to %s", container, repository+":"+tag), func() error {
		c.Logger.Info("commit container start")
		defer c.Logger.Info("commit container finished")

		if err := c.Docker.CommitContainerWithOptions(options); err != nil {
			c.Logger.Error(err)
			return err
		}
		return c.Refresh(g, v)
	})

	return nil
}

func (c *ContainerList) RenameContainerPanel(g *gocui.Gui, v *gocui.View) error {
	container, err := c.selected()
	if err != nil {
		c.ErrMessage(err.Error(), c.name)
		c.Logger.Error(err)
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

func (c *ContainerList) RenameContainer(g *gocui.Gui, v *gocui.View) error {
	if !c.form.Validate() {
		return nil
	}

	data := c.form.GetFieldTexts()
	oldName := data["Container"]
	name := data["NewName"]

	options := docker.RenameContainerOptions{
		ID:   oldName,
		Name: name,
	}

	c.form.Close(g, v)

	c.AddTask(fmt.Sprintf("Rename container %s to %s", oldName, name), func() error {
		c.Logger.Info("rename container start")
		defer c.Logger.Info("rename container finished")

		if err := c.Docker.RenameContainerWithOptions(options); err != nil {
			c.Logger.Error(err)
			return err
		}
		return c.Refresh(g, v)
	})

	return nil
}

func (c *ContainerList) Refresh(g *gocui.Gui, v *gocui.View) error {
	c.Update(func(g *gocui.Gui) error {
		v, err := c.View(c.name)
		if err != nil {
			c.Logger.Error(err)
			return nil
		}

		c.GetContainerList(v)

		return nil
	})

	return nil
}

func (c *ContainerList) GetContainerList(v *gocui.View) {
	v.Clear()
	c.Containers = make([]*Container, 0)

	for _, con := range c.Docker.Containers() {
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

		common.OutputFormatedLine(v, container)
	}
}

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
			c.Logger.Error(err)
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
		c.Logger.Error(err)
		return nil
	}

	return nil
}

func (c *ContainerList) AttachContainer(g *gocui.Gui, v *gocui.View) error {
	selected, err := c.selected()
	if err != nil {
		c.Logger.Error(err)
		return nil
	}

	container, err := c.Docker.InspectContainer(selected.ID)
	if err != nil {
		c.Logger.Error(err)
		return nil
	}

	if !container.State.Running {
		msg := fmt.Sprintf("container %s is not runnig", selected.Name)
		c.ErrMessage(msg, c.name)
		c.Logger.Error()
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
	form := NewForm(g, CreateContainerPanel, x, y, w, 0)
	c.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		c.SwitchPanel(c.name)
		return nil
	})

	attach := func(*gocui.Gui, *gocui.View) error {
		if !c.form.Validate() {
			return nil
		}
		return AttachFlag
	}

	// add fields
	form.AddInput("Cmd", labelw, fieldw).
		AddValidate("no specified Cmd", func(value string) bool {
			return value != ""
		})

	form.AddButton("Attach", attach)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()
	return nil
}

func (c *ContainerList) Attach() error {
	c.Logger.Info("attach container start")
	defer c.Logger.Info("attach container finished")

	selected, err := c.selected()
	if err != nil {
		c.Logger.Error(err)
		return nil
	}

	// https://gist.github.com/fsouza/43a05241ed9f943d24e5324c0f07471a
	fd := int(os.Stdin.Fd())
	if err != nil {
		c.Logger.Error(err)
		return err
	}

	if terminal.IsTerminal(fd) {
		oldState, err := terminal.MakeRaw(fd)
		if err != nil {
			c.Logger.Error(err)
			return err
		}
		defer terminal.Restore(fd, oldState)

		stdoutReader, stdoutWriter := io.Pipe()
		stderrReader, stderrWriter := io.Pipe()
		stdinReader, stdinWriter := io.Pipe()

		exec, err := c.Docker.CreateExec(docker.CreateExecOptions{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
			Cmd:          []string{c.form.GetFieldText("Cmd")},
			Container:    selected.ID,
		})

		if err != nil {
			c.Logger.Error(err)
			return err
		}

		waiter, err := c.Docker.StartExecNonBlocking(exec.ID, docker.StartExecOptions{
			InputStream:  stdinReader,
			OutputStream: stdoutWriter,
			ErrorStream:  stderrWriter,
			Tty:          true,
			RawTerminal:  true,
		})

		go io.Copy(stdinWriter, os.Stdin)
		go io.Copy(os.Stdout, stdoutReader)
		go io.Copy(os.Stderr, stderrReader)

		if err != nil {
			c.Logger.Error(err)
			return err
		}

		// reseize tty
		// https://github.com/fsouza/go-dockerclient/issues/771
		width, height, err := terminal.GetSize(fd)
		if err != nil {
			c.Logger.Error(err)
			return err
		}

		err = c.Docker.ResizeExecTTY(exec.ID, height, width)
		if err != nil {
			c.Logger.Error(err)
			return err
		}

		if err := waiter.Wait(); err != nil {
			c.Logger.Error(err)
		}
	} else {
		c.Logger.Error("no terminal")
		return nil
	}

	return nil
}
