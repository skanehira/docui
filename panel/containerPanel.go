package panel

import (
	"fmt"
	"log"
	"time"

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

func (i ContainerList) Name() string {
	return i.name
}

func (i ContainerList) SetView(g *gocui.Gui) (*gocui.View, error) {
	v, err := g.SetView(i.Name(), i.x, i.y, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}

		v.Title = v.Name()
		v.Wrap = true
		v.SetCursor(0, 1)

		return v, nil
	}

	return v, nil
}

func (i ContainerList) Init(g *Gui) {
	v, err := i.SetView(g.Gui)

	if err != nil {
		panic(err)
	}
	// keybinds
	g.SetKeybinds(i.Name())

	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlJ, gocui.ModNone, CursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlK, gocui.ModNone, CursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyEnter, gocui.ModNone, i.DetailContainer); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlD, gocui.ModNone, i.RemoveContainer); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlU, gocui.ModNone, i.StartContainer); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlS, gocui.ModNone, i.StopContainer); err != nil {
		log.Panicln(err)
	}

	i.LoadContainer(v)

	go func() {
		for {
			time.Sleep(5 * time.Second)
			i.LoadContainer(v)
		}
	}()

}

func (i ContainerList) DetailContainer(g *gocui.Gui, v *gocui.View) error {

	id := i.GetContainerID(v)
	if id == "" {
		return nil
	}

	container := i.Docker.InspectContainer(id)

	nv, err := g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	nv.Clear()
	nv.SetOrigin(0, 0)
	fmt.Fprint(nv, StructToJson(container))

	return nil
}

func (i ContainerList) RemoveContainer(g *gocui.Gui, v *gocui.View) error {
	id := i.GetContainerID(v)

	if id == "" {
		return nil
	}

	options := docker.RemoveContainerOptions{ID: id}
	if err := i.Docker.RemoveContainer(options); err != nil {
		i.DispMessage(err.Error(), ContainerListPanel)
		return nil
	}

	if err := i.RefreshPanel(g, v); err != nil {
		return err
	}

	return nil
}

func (i ContainerList) StartContainer(g *gocui.Gui, v *gocui.View) error {
	id := i.GetContainerID(v)
	if id == "" {
		return nil
	}

	if err := i.Docker.StartContainerWithID(id); err != nil {
		i.DispMessage(err.Error(), ContainerListPanel)
		return nil
	}

	if err := i.RefreshPanel(g, v); err != nil {
		return err
	}

	return nil
}

func (i ContainerList) StopContainer(g *gocui.Gui, v *gocui.View) error {
	id := i.GetContainerID(v)
	if id == "" {
		return nil
	}

	if err := i.Docker.StopContainerWithID(id); err != nil {
		i.DispMessage(err.Error(), ContainerListPanel)
		return nil
	}

	if err := i.RefreshPanel(g, v); err != nil {
		return err
	}

	return nil
}

func (i ContainerList) RefreshPanel(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		nv, err := g.View(ContainerListPanel)
		if err != nil {
			return err
		}

		v = nv
	}
	i.LoadContainer(v)
	SetCurrentPanel(g, v.Name())

	return nil
}

func (i ContainerList) LoadContainer(v *gocui.View) {
	v.Clear()
	fmt.Fprintf(v, "%-15s %-20s %-15s\n", "ID", "NAME", "STATUS")
	for _, c := range i.Docker.Containers() {
		fmt.Fprintf(v, "%-15s %-20s %-15s\n", c.ID[:12], c.Names[0][1:], c.Status)
	}
}

func (i ContainerList) GetContainerID(v *gocui.View) string {
	id := ReadLine(v, nil)
	if id == "" || id[:2] == "ID" {
		return ""
	}

	return id[:12]
}
