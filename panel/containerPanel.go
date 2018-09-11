package panel

import (
	"fmt"
	"log"
	"reflect"

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
		v.Autoscroll = true
		v.Wrap = true

		return v, nil
	}

	return v, nil
}

func (i ContainerList) Init(g *Gui) {
	v, err := i.SetView(g.Gui)

	if err != nil {
		panic(err)
	}

	i.LoadContainer(v)
	v.SetCursor(0, 1)

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
}

func (i ContainerList) DetailContainer(g *gocui.Gui, v *gocui.View) error {

	id := i.GetContainerID(v)
	if id == "" {
		return nil
	}

	img := i.Docker.InspectContainer(id)

	nv, err := g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	nv.Clear()

	value := reflect.Indirect(reflect.ValueOf(img))
	t := value.Type()

	// not display
	noDisplay := map[string]bool{
		"GraphDriver": true,
		"HostConfig":  true,
		"State":       true,
	}

	// display image detail
	for i := 0; i < t.NumField(); i++ {
		// field name
		fieldName := t.Field(i).Name

		if noDisplay[fieldName] {
			continue
		}

		fmt.Fprintf(nv, "%-15s: ", fieldName)

		value := value.Field(i)

		dispItem := func(i interface{}) {
			value := reflect.Indirect(reflect.ValueOf(i))
			t := value.Type()

			for i := 0; i < t.NumField(); i++ {
				// if i != 0 {
				// 	fmt.Fprintf(nv, "%-16s ", "")
				// }
				// field name
				fieldName := t.Field(i).Name
				if fieldName == "ExposedPorts" {
					fmt.Fprintf(nv, "%s: %v\n", fieldName, value.Field(i).Interface())
				}
				if fieldName == "Ports" {
					fmt.Fprintf(nv, "%s: %v\n", fieldName, value.Field(i).Interface())
				}
			}
		}

		switch fieldName {
		case "Config":
			c := value.Interface().(*docker.Config)
			dispItem(c)
			continue
		case "NetworkSettings":
			c := value.Interface().(*docker.NetworkSettings)
			dispItem(c)
			continue
		}

		fmt.Fprintf(nv, "%v\n", value.Interface())
	}

	return nil
}

func (i ContainerList) RemoveContainer(g *gocui.Gui, v *gocui.View) error {
	id := i.GetContainerID(v)

	if id == "" {
		return nil
	}

	options := docker.RemoveContainerOptions{ID: id}
	if err := i.Docker.RemoveContainer(options); err != nil {
		i.DispMessage(err.Error(), i)
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
		i.DispMessage(err.Error(), i)
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
		i.DispMessage(err.Error(), i)
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
	v.Clear()
	i.LoadContainer(v)
	SetCurrentPanel(g, v.Name())

	return nil
}

func (i ContainerList) LoadContainer(v *gocui.View) {
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
