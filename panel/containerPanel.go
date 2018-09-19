package panel

import (
	"fmt"
	"log"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

type ContainerList struct {
	*Gui
	name string
	Position
}

func NewContainerList(gui *Gui, name string, x, y, w, h int) ContainerList {
	return ContainerList{gui, name, Position{x, y, x + w, y + h}}
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
	c.SetKeybinds(c.name)

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
		options := docker.RemoveContainerOptions{ID: id}
		if err := c.Docker.RemoveContainer(options); err != nil {
			c.CloseConfirmMessage(g, v)
			c.ErrMessage(err.Error(), ContainerListPanel)
			return nil
		}
		c.CloseConfirmMessage(g, v)

		return nil
	})

	return nil
}

func (c ContainerList) StartContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)
	if id == "" {
		return nil
	}

	v = c.StateMessage("container starting...")
	g.Update(func(g *gocui.Gui) error {
		func(g *gocui.Gui, v *gocui.View) error {
			defer c.CloseStateMessage(v)
			if err := c.Docker.StartContainerWithID(id); err != nil {
				c.ErrMessage(err.Error(), ContainerListPanel)
				return nil
			}

			if err := c.Refresh(); err != nil {
				panic(err)
			}

			if _, err := SetCurrentPanel(g, ContainerListPanel); err != nil {
				panic(err)
			}

			return nil
		}(g, v)

		return nil
	})

	return nil
}

func (c ContainerList) StopContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)
	if id == "" {
		return nil
	}

	v = c.StateMessage("container stopping...")
	g.Update(func(g *gocui.Gui) error {
		func(g *gocui.Gui, v *gocui.View) error {
			defer c.CloseStateMessage(v)

			if err := c.Docker.StopContainerWithID(id); err != nil {
				c.ErrMessage(err.Error(), ContainerListPanel)
				return nil
			}

			if err := c.Refresh(); err != nil {
				panic(err)
			}

			if _, err := SetCurrentPanel(g, ContainerListPanel); err != nil {
				panic(err)
			}
			return nil
		}(g, v)

		return nil
	})

	return nil
}

func (c ContainerList) ExportContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)
	if id == "" {
		return nil
	}

	data := map[string]interface{}{
		"ID": id,
	}

	maxX, maxY := c.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	NewInput(c.Gui, ExportContainerPanel, x, y, w, h, NewExportContainerItems(x, y, w, h), data)
	return nil
}

func (i ContainerList) CommitContainer(g *gocui.Gui, v *gocui.View) error {
	id := i.GetContainerID(v)
	if id == "" {
		return nil
	}

	data := map[string]interface{}{
		"Container": id,
	}

	maxX, maxY := i.Size()
	x := maxX / 8
	y := maxY / 5
	w := maxX - x
	h := maxY - y

	NewInput(i.Gui, CommitContainerPanel, x, y, w, h, NewCommitContainerPanel(x, y, w, h), data)
	return nil
}

func (c ContainerList) Refresh() error {
	c.Update(func(g *gocui.Gui) error {
		if err := c.SetView(g); err != nil {
			return err
		}

		return nil
	})

	return nil
}

func (c ContainerList) GetContainerList(v *gocui.View) {
	v.Clear()
	fmt.Fprintf(v, "%-15s %-20s %-15s\n", "ID", "NAME", "STATUS")
	for _, con := range c.Docker.Containers() {
		fmt.Fprintf(v, "%-15s %-20s %-15s\n", con.ID[:12], con.Names[0][1:], con.Status)
	}
}

func (c ContainerList) GetContainerID(v *gocui.View) string {
	id := ReadLine(v, nil)
	if id == "" || id[:2] == "ID" {
		return ""
	}

	return id[:12]
}
