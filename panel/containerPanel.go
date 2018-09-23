package panel

import (
	"fmt"
	"log"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

type ContainerList struct {
	*Gui
	name string
	Position
	Containers map[string]Container
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
	return ContainerList{gui, name, Position{x, y, x + w, y + h}, make(map[string]Container)}
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
	// keybinds
	c.DeleteKeybindings(c.name)
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
	if err := c.SetKeybinding(c.name, 'e', gocui.ModNone, c.ExportContainer); err != nil {
		log.Panicln(err)
	}
	if err := c.SetKeybinding(c.name, 'c', gocui.ModNone, c.CommitContainer); err != nil {
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
	c.NextPanel = ContainerListPanel
	id := c.GetContainerID(v)

	if id == "" {
		return nil
	}

	c.ConfirmMessage("Do you want delete this container? (y/n)", func(g *gocui.Gui, v *gocui.View) error {
		defer c.Refresh()
		defer c.CloseConfirmMessage(g, v)
		options := docker.RemoveContainerOptions{ID: id}

		if err := c.Docker.RemoveContainer(options); err != nil {
			c.ErrMessage(err.Error(), ContainerListPanel)
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

	netxtPanel := ContainerListPanel

	g.Update(func(g *gocui.Gui) error {
		c.StateMessage("container starting...")

		g.Update(func(g *gocui.Gui) error {
			defer c.Refresh()
			defer c.CloseStateMessage()

			if err := c.Docker.StartContainerWithID(id); err != nil {
				c.ErrMessage(err.Error(), netxtPanel)
				return nil
			}

			c.SwitchPanel(netxtPanel)

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

	nextPanel := ContainerListPanel

	g.Update(func(g *gocui.Gui) error {
		c.StateMessage("container stopping...")

		g.Update(func(g *gocui.Gui) error {
			defer c.CloseStateMessage()
			defer c.Refresh()

			if err := c.Docker.StopContainerWithID(id); err != nil {
				c.ErrMessage(err.Error(), nextPanel)
				return nil
			}

			c.SwitchPanel(nextPanel)

			return nil
		})

		return nil
	})

	return nil
}

func (c ContainerList) ExportContainer(g *gocui.Gui, v *gocui.View) error {
	c.NextPanel = ContainerListPanel

	name := c.GetContainerName(v)
	if name == "" {
		return nil
	}

	data := map[string]interface{}{
		"Container": name,
	}

	maxX, maxY := c.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	NewInput(c.Gui, ExportContainerPanel, x, y, w, h, NewExportContainerItems(x, y, w, h), data)
	return nil
}

func (c ContainerList) CommitContainer(g *gocui.Gui, v *gocui.View) error {
	c.NextPanel = ContainerListPanel
	name := c.GetContainerName(v)
	if name == "" {
		return nil
	}

	data := map[string]interface{}{
		"Container": name,
	}

	maxX, maxY := c.Size()
	x := maxX / 8
	y := maxY / 3
	w := maxX - x
	h := maxY - y

	NewInput(c.Gui, CommitContainerPanel, x, y, w, h, NewCommitContainerPanel(x, y, w, h), data)
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

	format := "%-15s %-15s %-15s %-25s %-25s %-10s\n"
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

func NewCommitContainerPanel(ix, iy, iw, ih int) Items {
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
