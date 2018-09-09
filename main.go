package main

import (
	"fmt"
	"log"
	"reflect"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

// All panel name
const (
	ImageList     = "image list"
	ContainerList = "container list"
	Detail        = "detail"
	Message       = "message"
)

var (
	viewArr = []string{ImageList, ContainerList, Detail}
	active  = 0
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen

	// new docker
	d := NewDocker(endpoint)

	g.SetManagerFunc(d.layout)

	// keybinds
	if err := g.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(viewArr)
	name := viewArr[nextIndex]

	if _, err := setCurrentViewOnTop(g, name); err != nil {
		return err
	}

	active = nextIndex
	return nil
}

func (d *Docker) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView(ImageList, 0, 0, maxX/2-2, maxY/2-2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = v.Name()
		v.Autoscroll = true
		v.Wrap = true

		if _, err = setCurrentViewOnTop(g, ImageList); err != nil {
			return err
		}

		for _, i := range d.Images() {
			fmt.Fprintf(v, "%+v\n", i.RepoTags)
		}

		// keybinds
		if err := g.SetKeybinding(ImageList, gocui.KeyCtrlJ, gocui.ModNone, d.NextImage); err != nil {
			log.Panicln(err)
		}
		if err := g.SetKeybinding(ImageList, gocui.KeyCtrlK, gocui.ModNone, d.PreImage); err != nil {
			log.Panicln(err)
		}
		if err := g.SetKeybinding(ImageList, gocui.KeyEnter, gocui.ModNone, d.DetailImage); err != nil {
			log.Panicln(err)
		}
		if err := g.SetKeybinding(ImageList, gocui.KeyCtrlC, gocui.ModNone, d.CreateContainer); err != nil {
			log.Panicln(err)
		}
	}

	if v, err := g.SetView(ContainerList, 0, maxY/2-1, maxX/2-2, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = v.Name()
		v.Wrap = true
		v.Autoscroll = true

		for _, c := range d.Containers() {
			fmt.Fprintf(v, "name:%+v status:%s\n", c.Names, c.Status)
		}
	}

	if v, err := g.SetView(Detail, maxX/2-1, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = v.Name()
		v.Wrap = true
		v.Autoscroll = true
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func ClearView(g *gocui.Gui, views ...string) {
	for _, v := range views {
		g.DeleteView(v)
	}
}

func (d *Docker) NextImage(g *gocui.Gui, v *gocui.View) error {
	ClearView(g, Detail)

	x, y := v.Cursor()
	nextY := y + 1
	str, _ := v.Line(nextY)

	if str != "" {
		v.SetCursor(x, nextY)
	}

	return nil
}

func (d *Docker) PreImage(g *gocui.Gui, v *gocui.View) error {
	ClearView(g, Detail)

	x, y := v.Cursor()
	preY := y - 1
	str, _ := v.Line(preY)

	if str != "" {
		v.SetCursor(x, preY)
	}

	return nil
}

func (d *Docker) DetailImage(g *gocui.Gui, v *gocui.View) error {
	_, y := v.Cursor()

	imgName, _ := v.Line(y)
	imgName = imgName[1 : len(imgName)-1]

	img := d.InspectImage(imgName)

	nv, err := g.View(Detail)
	if err != nil {
		panic(err)
	}
	nv.Clear()

	value := reflect.Indirect(reflect.ValueOf(img))
	t := value.Type()

	// not display
	noDisplay := map[string]bool{
		"RootFS":      true,
		"RepoDigests": true,
		"Config":      true,
		//"ContainerConfig": true,
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

		if fieldName == "ContainerConfig" {
			c := value.Interface().(docker.Config)

			value := reflect.Indirect(reflect.ValueOf(c))
			t := value.Type()

			for i := 0; i < t.NumField(); i++ {
				// field name
				fieldName := t.Field(i).Name
				if i != 0 {
					fmt.Fprintf(nv, "%-16s ", "")
				}
				fmt.Fprintf(nv, "%s: %v\n", fieldName, value.Field(i).Interface())
			}
			continue
		}

		fmt.Fprintf(nv, "%v\n", value.Interface())
	}

	return nil
}

func (d *Docker) CreateContainer(g *gocui.Gui, v *gocui.View) error {

	_, y := v.Cursor()

	imgName, _ := v.Line(y)
	imgName = imgName[1 : len(imgName)-1]

	items := &Items{
		items: map[int]Item{
			0: Item{disp: "Name", name: Position{0, 15, 0}, input: Position{16, 40, 0}},
			1: Item{disp: "HostPort", name: Position{0, 15, 4}, input: Position{16, 40, 4}},
			2: Item{disp: "Port", name: Position{0, 15, 8}, input: Position{16, 40, 8}},
			3: Item{disp: "HostVolume", name: Position{0, 15, 12}, input: Position{16, 40, 12}},
			4: Item{disp: "Volume", name: Position{0, 15, 16}, input: Position{16, 40, 16}},
			5: Item{disp: "Image", name: Position{0, 15, 20}, input: Position{16, 40, 20}},
			6: Item{disp: "Env", name: Position{0, 15, 24}, input: Position{16, 40, 24}},
			7: Item{disp: "Enter", name: Position{0, 15, 28}, input: Position{16, 40, 28}},
		},
	}
	// disable change panel
	if err := g.DeleteKeybinding("", gocui.KeyTab, gocui.ModNone); err != nil {
		log.Panicln(err)
	}

	if err := d.InputPanel(g, "create container", items, imgName); err != nil {
		return err
	}

	return nil
}

type Position struct {
	startX, endX, y int
}

type Item struct {
	disp        string
	name, input Position
}

type Items struct {
	items map[int]Item
}

func (d *Docker) InputPanel(g *gocui.Gui, viewName string, items *Items, img string) error {
	// clear view
	for _, item := range items.items {
		ClearView(g, item.disp)
	}

	ClearView(g, viewName)

	// create cntainer view
	maxX, maxY := g.Size()
	x0 := maxX / 6
	y0 := maxY / 6
	v, err := g.SetView(viewName, x0, y0, maxX-x0, maxY-y0)

	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = viewName
		v.Wrap = true
		v.Editable = true
	}

	// create input
	views := make(map[string]*gocui.View)

	for i := 0; i < len(items.items); i++ {
		item := items.items[i]

		if item.disp != "Enter" {
			v.SetCursor(item.name.startX+2, item.name.y+1)
			for _, name := range []rune(item.disp) {
				v.EditWrite(name)
			}
		}

		if v, err := g.SetView(item.disp, x0+item.input.startX, y0+item.input.y+1, x0+item.input.endX, y0+item.input.y+3); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			} else {
				if item.disp == "Enter" {
					v.SetCursor(9, 0)
					for _, name := range []rune(item.disp) {
						v.EditWrite(name)
					}
					continue
				}

				if item.disp == "Image" {
					v.SetCursor(0, 0)
					for _, name := range []rune(img) {
						v.EditWrite(name)
					}
				}

				v.Editable = true
				v.Wrap = true
				views[item.disp] = v
			}
		}
	}

	item := items.items[0]
	setCurrentViewOnTop(g, item.disp)
	views[item.disp].SetCursor(item.input.startX, item.input.y+1)

	// select input keybinds
	active := 0

	next := func(g *gocui.Gui, v *gocui.View) error {
		if active == len(items.items)-1 {
			return nil
		}

		nextIndex := active + 1
		name := items.items[nextIndex].disp

		if _, err := setCurrentViewOnTop(g, name); err != nil {
			return err
		}

		active = nextIndex
		return nil
	}

	pre := func(g *gocui.Gui, v *gocui.View) error {
		if active == 0 {
			return nil
		}

		nextIndex := active - 1
		name := items.items[nextIndex].disp

		if _, err := setCurrentViewOnTop(g, name); err != nil {
			return err
		}

		active = nextIndex
		return nil
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlJ, gocui.ModNone, next); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlK, gocui.ModNone, pre); err != nil {
		log.Panicln(err)
	}

	// when hit enter, create container
	createContainer := func(g *gocui.Gui, v *gocui.View) error {
		config := make(map[string]string)
		for name, v := range views {
			value, _ := v.Line(0)
			config[name] = value
		}

		if err := d.CreateContainerWithOptions(config); err != nil {
			PopUp(g, err.Error(), viewName)
			if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
				log.Panicln(err)
			}
			return nil
		}

		PopUp(g, "create container success", ContainerList)
		if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
			log.Panicln(err)
		}

		return nil
	}

	if err := g.SetKeybinding("Enter", gocui.KeyEnter, gocui.ModNone, createContainer); err != nil {
		log.Panicln(err)
	}

	return nil
}

func PopUp(g *gocui.Gui, message, nextView string) error {
	maxX, maxY := g.Size()
	x0 := maxX / 4
	y0 := maxY / 4
	viewName := "message"

	if v, err := g.SetView(viewName, x0, y0, maxX-x0, y0+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		fmt.Fprint(v, message)
		if _, err := setCurrentViewOnTop(g, viewName); err != nil {
			return err
		}

		close := func(g *gocui.Gui, v *gocui.View) error {
			g.SetViewOnBottom(viewName)
			if _, err := setCurrentViewOnTop(g, nextView); err != nil {
				return err
			}
			return nil
		}

		if err := g.SetKeybinding(viewName, gocui.KeyEnter, gocui.ModNone, close); err != nil {
			log.Panicln(err)
		}
	}

	return nil
}
