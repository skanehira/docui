package panel

import (
	"fmt"
	"log"
	"reflect"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

type ImageList struct {
	*Gui
	name string
	Position
}

func NewImageList(gui *Gui, name string, x, y, w, h int) ImageList {
	return ImageList{gui, name, Position{x, y, x + w, y + h}}
}

func (i ImageList) Name() string {
	return i.name
}

func (i ImageList) SetView(g *gocui.Gui) (*gocui.View, error) {
	v, err := g.SetView(i.Name(), i.x, i.y, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}

		v.Title = v.Name()
		v.Autoscroll = true
		v.Wrap = true

		if _, err = SetCurrentPanel(g, i.Name()); err != nil {
			return nil, err
		}

		return v, nil
	}

	return v, nil
}

func (i ImageList) Init(g *Gui) {
	v, err := i.SetView(g.Gui)

	if err != nil {
		panic(err)
	}

	for _, i := range g.Docker.Images() {
		fmt.Fprintf(v, "%+v\n", i.RepoTags)
	}

	// keybinds
	g.SetKeybinds(i.Name())

	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlJ, gocui.ModNone, CursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlK, gocui.ModNone, CursorUp); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(i.Name(), gocui.KeyEnter, gocui.ModNone, i.DetailImage); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlC, gocui.ModNone, i.CreateContainer); err != nil {
		log.Panicln(err)
	}

}

func (i ImageList) CreateContainer(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()
	imgName := ReadLine(v, nil)
	if imgName == "" {
		return nil
	}

	data := map[string]interface{}{
		"Image": imgName,
	}

	input := NewInput(i.Gui, CreateContainerPanel, maxX/8, maxY/8, maxX-maxX/4-2, maxY-maxY/4-2, 40, NewCreateContainerItems(), data)
	input.Init(i.Gui)
	return nil
}

func (i ImageList) DetailImage(g *gocui.Gui, v *gocui.View) error {

	imgName := ReadLine(v, nil)
	if imgName == "" {
		return nil
	}

	imgName = imgName[1 : len(imgName)-1]

	img := i.Docker.InspectImage(imgName)

	nv, err := g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	nv.Clear()
	nv.SetCursor(0, 0)

	value := reflect.Indirect(reflect.ValueOf(img))
	t := value.Type()

	// not display
	noDisplay := map[string]bool{
		"RootFS":      true,
		"RepoDigests": true,
		"Config":      true,
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

		dispConfig := func(i interface{}) {
			value := reflect.Indirect(reflect.ValueOf(i))
			t := value.Type()

			for i := 0; i < t.NumField(); i++ {
				// field name
				fieldName := t.Field(i).Name
				if i != 0 {
					fmt.Fprintf(nv, "%-16s ", "")
				}
				fmt.Fprintf(nv, "%s: %v\n", fieldName, value.Field(i).Interface())
			}
		}

		switch fieldName {
		case "ContainerConfig":
			c := value.Interface().(docker.Config)
			dispConfig(c)
			continue

		case "ID":
			fmt.Fprintf(nv, "%s\n", value.String()[7:])
			continue
			// case "Config":
			// c := value.Interface().(*docker.Config)
			// dispConfig(c)
			// continue
		}

		fmt.Fprintf(nv, "%v\n", value.Interface())
	}

	return nil
}
