package panel

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

type ContainerList struct {
	*Gui
	name string
	Position
	Containers     map[string]Container
	Data           map[string]interface{}
	ClosePanelName string
}

type Container struct {
	ID      string
	Image   string
	Created string
	Status  string
	Port    string
	Name    string
}

func NewContainerList(gui *Gui, name string, x, y, w, h int) ContainerList {
	return ContainerList{
		Gui:        gui,
		name:       name,
		Position:   Position{x, y, x + w, y + h},
		Containers: make(map[string]Container),
		Data:       make(map[string]interface{}),
	}
}

func (c ContainerList) Name() string {
	return c.name
}

func (c ContainerList) SetView(g *gocui.Gui) error {
	v, err := g.SetView(c.name, c.x, c.y, c.w, c.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = v.Name()
		v.Wrap = true
		v.SetOrigin(0, 0)
		v.SetCursor(0, 1)
	}

	c.SetKeyBinding()
	c.GetContainerList(v)

	return nil
}

func (c ContainerList) SetKeyBinding() {
	c.SetKeyBindingToPanel(c.name)

	if err := c.SetKeybinding(c.name, 'j', gocui.ModNone, CursorDown); err != nil {
		log.Panicln(err)
	}
	if err := c.SetKeybinding(c.name, 'k', gocui.ModNone, CursorUp); err != nil {
		log.Panicln(err)
	}
	if err := c.SetKeybinding(c.name, gocui.KeyEnter, gocui.ModNone, c.DetailContainer); err != nil {
		log.Panicln(err)
	}
	if err := c.SetKeybinding(c.name, 'o', gocui.ModNone, c.DetailContainer); err != nil {
		log.Panicln(err)
	}
	if err := c.SetKeybinding(c.name, 'd', gocui.ModNone, c.RemoveContainer); err != nil {
		log.Panicln(err)
	}
	if err := c.SetKeybinding(c.name, 'u', gocui.ModNone, c.StartContainer); err != nil {
		log.Panicln(err)
	}
	if err := c.SetKeybinding(c.name, 's', gocui.ModNone, c.StopContainer); err != nil {
		log.Panicln(err)
	}
	if err := c.SetKeybinding(c.name, 'e', gocui.ModNone, c.ExportContainerPanel); err != nil {
		log.Panicln(err)
	}
	if err := c.SetKeybinding(c.name, 'c', gocui.ModNone, c.CommitContainerPanel); err != nil {
		log.Panicln(err)
	}
	if err := c.SetKeybinding(c.name, 'r', gocui.ModNone, c.RenameContainerPanel); err != nil {
		log.Panicln(err)
	}
}

func (c ContainerList) DetailContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)
	if id == "" {
		return nil
	}

	container := c.Docker.InspectContainer(id)

	v, err := g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)
	fmt.Fprint(v, StructToJson(container))

	return nil
}

func (c ContainerList) RemoveContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)

	if id == "" {
		return nil
	}

	c.NextPanel = ContainerListPanel

	c.ConfirmMessage("Do you want delete this container? (y/n)", func(g *gocui.Gui, v *gocui.View) error {
		defer c.Refresh()
		defer c.CloseConfirmMessage(g, v)
		options := docker.RemoveContainerOptions{ID: id}

		if err := c.Docker.RemoveContainer(options); err != nil {
			c.ErrMessage(err.Error(), c.NextPanel)
			return nil
		}

		return nil
	})

	return nil
}

func (c ContainerList) StartContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)
	if id == "" {
		return nil
	}

	c.NextPanel = ContainerListPanel

	g.Update(func(g *gocui.Gui) error {
		c.StateMessage("container starting...")

		g.Update(func(g *gocui.Gui) error {
			defer c.Refresh()
			defer c.CloseStateMessage()

			if err := c.Docker.StartContainerWithID(id); err != nil {
				c.ErrMessage(err.Error(), c.NextPanel)
				return nil
			}

			c.SwitchPanel(c.NextPanel)

			return nil
		})

		return nil
	})

	return nil
}

func (c ContainerList) StopContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)
	if id == "" {
		return nil
	}

	c.NextPanel = ContainerListPanel

	g.Update(func(g *gocui.Gui) error {
		c.StateMessage("container stopping...")

		g.Update(func(g *gocui.Gui) error {
			defer c.CloseStateMessage()
			defer c.Refresh()

			if err := c.Docker.StopContainerWithID(id); err != nil {
				c.ErrMessage(err.Error(), c.NextPanel)
				return nil
			}

			c.SwitchPanel(c.NextPanel)

			return nil
		})

		return nil
	})

	return nil
}

func (c ContainerList) ExportContainerPanel(g *gocui.Gui, v *gocui.View) error {

	name := c.GetContainerName(v)
	if name == "" {
		return nil
	}

	c.Data = map[string]interface{}{
		"Container": name,
	}

	maxX, maxY := c.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	c.NextPanel = ContainerListPanel
	c.ClosePanelName = ExportContainerPanel

	handlers := Handlers{
		gocui.KeyEnter: c.ExportContainer,
	}

	NewInput(c.Gui, ExportContainerPanel, x, y, w, h, NewExportContainerItems(x, y, w, h), c.Data, handlers)
	return nil
}

func (c ContainerList) ExportContainer(g *gocui.Gui, v *gocui.View) error {
	path := ReadLine(v, nil)

	if path == "" {
		return nil
	}

	g.Update(func(g *gocui.Gui) error {
		c.ClosePanel(g, v)
		c.StateMessage("container exporting...")

		g.Update(func(g *gocui.Gui) error {
			defer c.CloseStateMessage()

			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
			if err != nil {
				c.ErrMessage(err.Error(), c.NextPanel)
				return nil
			}
			defer file.Close()

			options := docker.ExportContainerOptions{
				ID:           c.Data["Container"].(string),
				OutputStream: file,
			}

			if err := c.Docker.ExportContainerWithOptions(options); err != nil {
				c.ErrMessage(err.Error(), c.NextPanel)
				return nil
			}

			c.SwitchPanel(c.NextPanel)

			return nil

		})
		return nil
	})

	return nil
}

func (c ContainerList) CommitContainerPanel(g *gocui.Gui, v *gocui.View) error {
	name := c.GetContainerName(v)
	if name == "" {
		return nil
	}

	c.Data = map[string]interface{}{
		"Container": name,
	}

	maxX, maxY := c.Size()
	x := maxX / 8
	y := maxY / 3
	w := maxX - x
	h := maxY - y

	c.ClosePanelName = CommitContainerPanel
	c.NextPanel = ContainerListPanel

	handlers := Handlers{
		gocui.KeyEnter: c.CommitContainer,
	}

	NewInput(c.Gui, CommitContainerPanel, x, y, w, h, NewCommitContainerItems(x, y, w, h), c.Data, handlers)
	return nil
}

func (c ContainerList) CommitContainer(g *gocui.Gui, v *gocui.View) error {

	data, err := c.GetItemsToMap(NewCommitContainerItems(c.x, c.y, c.w, c.h))
	if err != nil {
		c.ClosePanel(g, v)
		c.ErrMessage(err.Error(), c.NextPanel)
		return nil
	}

	options := docker.CommitContainerOptions{
		Container:  data["Container"],
		Repository: data["Repository"],
		Tag:        data["Tag"],
	}

	g.Update(func(g *gocui.Gui) error {
		c.ClosePanel(g, v)
		c.StateMessage("container committing...")

		g.Update(func(g *gocui.Gui) error {
			defer c.CloseStateMessage()

			if err := c.Docker.CommitContainerWithOptions(options); err != nil {
				c.ErrMessage(err.Error(), c.NextPanel)
				return nil
			}

			c.Panels[ImageListPanel].Refresh()
			c.SwitchPanel(c.NextPanel)

			return nil

		})

		return nil
	})

	return nil
}

func (c ContainerList) RenameContainerPanel(g *gocui.Gui, v *gocui.View) error {

	name := c.GetContainerName(v)
	if name == "" {
		return nil
	}

	c.Data = map[string]interface{}{
		"Container": name,
	}

	maxX, maxY := c.Size()
	x := maxX / 8
	y := maxY/3 + 5
	w := maxX - x
	h := maxY - y

	c.ClosePanelName = RenameContainerPanel
	c.NextPanel = ContainerListPanel

	handlers := Handlers{
		gocui.KeyEnter: c.RenameContainer,
	}

	NewInput(c.Gui, RenameContainerPanel, x, y, w, h, NewRenameContainerItems(x, y, w, h), c.Data, handlers)
	return nil
}

func (c ContainerList) RenameContainer(g *gocui.Gui, v *gocui.View) error {

	data, err := c.GetItemsToMap(NewRenameContainerItems(c.x, c.y, c.w, c.h))
	if err != nil {
		c.ClosePanel(g, v)
		c.ErrMessage(err.Error(), c.NextPanel)
		return nil
	}

	options := docker.RenameContainerOptions{
		ID:   data["Container"],
		Name: data["NewName"],
	}

	g.Update(func(g *gocui.Gui) error {
		c.ClosePanel(g, v)
		c.StateMessage("container renaming...")

		g.Update(func(g *gocui.Gui) error {
			defer c.CloseStateMessage()

			if err := c.Docker.RenameContainerWithOptions(options); err != nil {
				c.ErrMessage(err.Error(), c.NextPanel)
				return nil
			}

			c.Refresh()
			c.SwitchPanel(c.NextPanel)

			return nil

		})

		return nil
	})

	return nil
}

func (c ContainerList) Refresh() error {
	c.Update(func(g *gocui.Gui) error {
		v, err := c.View(ContainerListPanel)
		if err != nil {
			panic(err)
		}

		c.GetContainerList(v)

		return nil
	})

	return nil
}

func (c ContainerList) GetContainerList(v *gocui.View) {
	v.Clear()

	c1, c2, c3, c4, c5, c6 := 15, 15, 15, 15, 25, 25

	format := "%-" + strconv.Itoa(c1) + "s %-" + strconv.Itoa(c2) + "s %-" + strconv.Itoa(c3) + "s %-" + strconv.Itoa(c4) + "s %-" + strconv.Itoa(c5) + "s %-" + strconv.Itoa(c6) + "s\n"
	fmt.Fprintf(v, format, "ID", "IMAGE", "NAME", "STATUS", "CREATED", "PORT")

	for _, con := range c.Docker.Containers() {
		id := con.ID[:12]
		image := con.Image
		name := con.Names[0][1:]
		status := con.Status
		created := ParseDateToString(con.Created)
		port := ParsePortToString(con.Ports)

		c.Containers[id] = Container{
			ID:      con.ID,
			Image:   image,
			Name:    name,
			Status:  status,
			Created: created,
			Port:    port,
		}

		if len(image) > c2 {
			image = image[:c2-3] + "..."
		}
		if len(name) > c3 {
			name = name[:c3-3] + "..."
		}
		if len(status) > c4 {
			status = status[:c4-3] + "..."
		}
		if len(created) > c5 {
			created = created[:c5-3] + "..."
		}
		if len(port) > c6 {
			port = port[:c6-3] + "..."
		}

		fmt.Fprintf(v, format, id, image, name, status, created, port)
	}
}

func (c ContainerList) GetContainerID(v *gocui.View) string {
	line := ReadLine(v, nil)
	if line == "" {
		return line
	}

	return strings.Split(line, " ")[0]
}

func (c ContainerList) GetContainerName(v *gocui.View) string {
	return c.Containers[c.GetContainerID(v)].Name
}

func (c ContainerList) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	return c.Panels[c.ClosePanelName].(Input).ClosePanel(g, v)
}

func NewCommitContainerItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Repository",
		"Tag",
		"Container",
	}

	return NewItems(names, ix, iy, iw, ih, 12)
}

func NewCreateContainerItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Name",
		"HostPort",
		"Port",
		"HostVolume",
		"Volume",
		"Image",
		"Env",
		"Cmd",
	}

	return NewItems(names, ix, iy, iw, ih, 12)
}

func NewExportContainerItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Path",
	}

	return NewItems(names, ix, iy, iw, ih, 6)
}

func NewRenameContainerItems(ix, iy, iw, ih int) Items {
	names := []string{
		"NewName",
		"Container",
	}

	return NewItems(names, ix, iy, iw, ih, 12)
}
