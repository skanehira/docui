package panel

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

type ContainerList struct {
	*Gui
	name string
	Position
	Containers     map[string]*Container
	Data           map[string]interface{}
	ClosePanelName string
	Items          Items
}

type Container struct {
	ID      string `tag:"ID" len:"min:15 max:0.1"`
	Name    string `tag:"NAME" len:"min:20 max:0.2"`
	Image   string `tag:"IMAGE" len:"min:20 max:0.2"`
	Status  string `tag:"STATUS" len:"min:15 max:0.2"`
	Created string `tag:"CREATED" len:"min:20 max:0.1"`
	Port    string `tag:"PORT" len:"min:20 max:0.2"`
}

func NewContainerList(gui *Gui, name string, x, y, w, h int) *ContainerList {
	return &ContainerList{
		Gui:        gui,
		name:       name,
		Position:   Position{x, y, w, h},
		Containers: make(map[string]*Container),
		Data:       make(map[string]interface{}),
		Items:      Items{},
	}
}

func (c *ContainerList) Name() string {
	return c.name
}

func (c *ContainerList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := g.SetView(ContainerListHeaderPanel, c.x, c.y, c.w, c.h); err != nil {
		if err != gocui.ErrUnknownView {
			panic(err)
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
			return err
		}

		v.Frame = false
		v.Wrap = true
		v.FgColor = gocui.ColorGreen
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)
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
	if err := c.SetKeybinding(c.name, gocui.KeyCtrlR, gocui.ModNone, c.Refresh); err != nil {
		log.Panicln(err)
	}
}

func (c *ContainerList) DetailContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)
	if id == "" {
		return nil
	}

	container := c.Docker.InspectContainer(id)

	c.PopupDetailPanel(g, v)

	v, err := g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)
	fmt.Fprint(v, common.StructToJson(container))

	return nil
}

func (c *ContainerList) RemoveContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)

	if id == "" {
		return nil
	}

	c.NextPanel = ContainerListPanel

	c.ConfirmMessage("Are you sure you want to remove this container? (y/n)", func(g *gocui.Gui, v *gocui.View) error {
		defer c.Refresh(g, v)
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

func (c *ContainerList) StartContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)
	if id == "" {
		return nil
	}

	c.NextPanel = ContainerListPanel

	g.Update(func(g *gocui.Gui) error {
		c.StateMessage("container starting...")

		g.Update(func(g *gocui.Gui) error {
			defer c.Refresh(g, v)
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

func (c *ContainerList) StopContainer(g *gocui.Gui, v *gocui.View) error {
	id := c.GetContainerID(v)
	if id == "" {
		return nil
	}

	c.NextPanel = ContainerListPanel

	g.Update(func(g *gocui.Gui) error {
		c.StateMessage("container stopping...")

		g.Update(func(g *gocui.Gui) error {
			defer c.CloseStateMessage()
			defer c.Refresh(g, v)

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

func (c *ContainerList) ExportContainerPanel(g *gocui.Gui, v *gocui.View) error {

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
	h := y + 4

	c.NextPanel = ContainerListPanel
	c.ClosePanelName = ExportContainerPanel
	c.Items = c.NewExportContainerItems(x, y, w, h)

	handlers := Handlers{
		gocui.KeyEnter: c.ExportContainer,
	}

	NewInput(c.Gui, ExportContainerPanel, x, y, w, h, c.Items, c.Data, handlers)
	return nil
}

func (c *ContainerList) ExportContainer(g *gocui.Gui, v *gocui.View) error {
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

func (c *ContainerList) CommitContainerPanel(g *gocui.Gui, v *gocui.View) error {
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
	c.Items = c.NewCommitContainerItems(x, y, w, h)

	handlers := Handlers{
		gocui.KeyEnter: c.CommitContainer,
	}

	NewInput(c.Gui, CommitContainerPanel, x, y, w, h, c.Items, c.Data, handlers)
	return nil
}

func (c *ContainerList) CommitContainer(g *gocui.Gui, v *gocui.View) error {

	data, err := c.GetItemsToMap(c.Items)
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

			c.Panels[ImageListPanel].Refresh(g, v)
			c.SwitchPanel(c.NextPanel)

			return nil

		})

		return nil
	})

	return nil
}

func (c *ContainerList) RenameContainerPanel(g *gocui.Gui, v *gocui.View) error {

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
	c.Items = c.NewRenameContainerItems(x, y, w, h)

	handlers := Handlers{
		gocui.KeyEnter: c.RenameContainer,
	}

	NewInput(c.Gui, RenameContainerPanel, x, y, w, h, c.Items, c.Data, handlers)
	return nil
}

func (c *ContainerList) RenameContainer(g *gocui.Gui, v *gocui.View) error {

	data, err := c.GetItemsToMap(c.Items)
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

			c.Refresh(g, v)
			c.SwitchPanel(c.NextPanel)

			return nil

		})

		return nil
	})

	return nil
}

func (c *ContainerList) Refresh(g *gocui.Gui, v *gocui.View) error {
	c.Update(func(g *gocui.Gui) error {
		v, err := c.View(c.name)
		if err != nil {
			panic(err)
		}

		c.GetContainerList(v)

		return nil
	})

	return nil
}

func (c *ContainerList) GetContainerList(v *gocui.View) {
	v.Clear()

	for _, con := range c.Docker.Containers() {
		id := con.ID[:12]
		image := con.Image
		name := con.Names[0][1:]
		status := con.Status
		created := ParseDateToString(con.Created)
		port := ParsePortToString(con.Ports)

		container := &Container{
			ID:      id,
			Image:   image,
			Name:    name,
			Status:  status,
			Created: created,
			Port:    port,
		}

		c.Containers[id] = container

		common.OutputFormatedLine(v, container)
	}
}

func (c *ContainerList) GetContainerID(v *gocui.View) string {
	line := ReadLine(v, nil)
	if line == "" {
		return ""
	}

	return strings.Split(line, " ")[0]
}

func (c *ContainerList) GetContainerName(v *gocui.View) string {
	return c.Containers[c.GetContainerID(v)].Name
}

func (c *ContainerList) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	return c.Panels[c.ClosePanelName].(*Input).ClosePanel(g, v)
}

func (c *ContainerList) NewCommitContainerItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Repository",
		"Tag",
		"Container",
	}

	return NewItems(names, ix, iy, iw, ih, 12)
}

func (c *ContainerList) NewExportContainerItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Path",
	}

	return NewItems(names, ix, iy, iw, ih, 6)
}

func (c *ContainerList) NewRenameContainerItems(ix, iy, iw, ih int) Items {
	names := []string{
		"NewName",
		"Container",
	}

	return NewItems(names, ix, iy, iw, ih, 12)
}
