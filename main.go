package main

import (
	"docui/panel"

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
	gui := panel.New(gocui.Output256)
	defer gui.Close()

	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}

//
//func ClearView(g *gocui.Gui, views ...string) {
//	for _, v := range views {
//		g.DeleteView(v)
//	}
//}
//
//
//
//type position struct {
//	x, y, w, h int
//}
//
//type Item struct {
//	label map[string]position
//	input map[string]position
//}
//
//type Items map[int]Item
//
//func NewItems(items []string, view string) Items {
//
//}
//
//func (d *Docker) CreateContainer(g *gocui.Gui, v *gocui.View) error {
//
//	_, y := v.Cursor()
//
//	imgName, _ := v.Line(y)
//	imgName = imgName[1 : len(imgName)-1]
//
//	items := &Items{
//		0: Item{label: "Name", name: Position{0, 15, 0}, input: Position{16, 40, 0}},
//		1: Item{label: "HostPort", name: Position{0, 15, 4}, input: Position{16, 40, 4}},
//		2: Item{label: "Port", name: Position{0, 15, 8}, input: Position{16, 40, 8}},
//		3: Item{label: "HostVolume", name: Position{0, 15, 12}, input: Position{16, 40, 12}},
//		4: Item{label: "Volume", name: Position{0, 15, 16}, input: Position{16, 40, 16}},
//		5: Item{label: "Image", name: Position{0, 15, 20}, input: Position{16, 40, 20}},
//		6: Item{label: "Env", name: Position{0, 15, 24}, input: Position{16, 40, 24}},
//		7: Item{label: "Enter", name: Position{0, 15, 28}, input: Position{16, 40, 28}},
//	}
//	// disable change panel
//	if err := g.DeleteKeybinding("", gocui.KeyTab, gocui.ModNone); err != nil {
//		log.Panicln(err)
//	}
//
//	if err := d.InputPanel(g, "create container", items, imgName); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (d *Docker) InputPanel(g *gocui.Gui, viewName string, items *Items, img string) error {
//	// clear view
//	for _, item := range items.items {
//		ClearView(g, item.label)
//	}
//
//	ClearView(g, viewName)
//
//	// create cntainer view
//	maxX, maxY := g.Size()
//	x0 := maxX / 6
//	y0 := maxY / 6
//	v, err := g.SetView(viewName, x0, y0, maxX-x0, maxY-y0)
//
//	if err != nil {
//		if err != gocui.ErrUnknownView {
//			return err
//		}
//		v.Title = viewName
//		v.Wrap = true
//		v.Editable = true
//	}
//
//	// create input
//	views := make(map[string]*gocui.View)
//
//	for i := 0; i < len(items.items); i++ {
//		item := items.items[i]
//
//		if item.label != "Enter" {
//			v.SetCursor(item.name.x+2, item.name.y+1)
//			for _, name := range []rune(item.label) {
//				v.EditWrite(name)
//			}
//		}
//
//		if v, err := g.SetView(item.label, x0+item.input.x, y0+item.input.y+1, x0+item.input.y, y0+item.input.y+3); err != nil {
//			if err != gocui.ErrUnknownView {
//				return err
//			} else {
//				if item.label == "Enter" {
//					v.SetCursor(9, 0)
//					for _, name := range []rune(item.label) {
//						v.EditWrite(name)
//					}
//					continue
//				}
//
//				if item.label == "Image" {
//					v.SetCursor(0, 0)
//					for _, name := range []rune(img) {
//						v.EditWrite(name)
//					}
//				}
//
//				v.Editable = true
//				v.Wrap = true
//				views[item.label] = v
//			}
//		}
//	}
//
//	item := items.items[0]
//	setCurrentViewOnTop(g, item.label)
//	views[item.label].SetCursor(item.input.x, item.input.y+1)
//
//	// select input keybinds
//	active := 0
//
//	next := func(g *gocui.Gui, v *gocui.View) error {
//		if active == len(items.items)-1 {
//			return nil
//		}
//
//		nextIndex := active + 1
//		name := items.items[nextIndex].label
//
//		if _, err := setCurrentViewOnTop(g, name); err != nil {
//			return err
//		}
//
//		active = nextIndex
//		return nil
//	}
//
//	pre := func(g *gocui.Gui, v *gocui.View) error {
//		if active == 0 {
//			return nil
//		}
//
//		nextIndex := active - 1
//		name := items.items[nextIndex].label
//
//		if _, err := setCurrentViewOnTop(g, name); err != nil {
//			return err
//		}
//
//		active = nextIndex
//		return nil
//	}
//
//	if err := g.SetKeybinding("", gocui.KeyCtrlJ, gocui.ModNone, next); err != nil {
//		log.Panicln(err)
//	}
//
//	if err := g.SetKeybinding("", gocui.KeyCtrlK, gocui.ModNone, pre); err != nil {
//		log.Panicln(err)
//	}
//
//	// when hit enter, create container
//	createContainer := func(g *gocui.Gui, v *gocui.View) error {
//		config := make(map[string]string)
//		for name, v := range views {
//			value, _ := v.Line(0)
//			config[name] = value
//		}
//
//		if err := d.CreateContainerWithOptions(config); err != nil {
//			PopUp(g, err.Error(), viewName)
//			if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
//				log.Panicln(err)
//			}
//			return nil
//		}
//
//		PopUp(g, "create container success", ContainerList)
//		if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
//			log.Panicln(err)
//		}
//
//		return nil
//	}
//
//	if err := g.SetKeybinding("Enter", gocui.KeyEnter, gocui.ModNone, createContainer); err != nil {
//		log.Panicln(err)
//	}
//
//	return nil
//}
//
//func PopUp(g *gocui.Gui, message, nextView string) error {
//	maxX, maxY := g.Size()
//	x0 := maxX / 4
//	y0 := maxY / 4
//	viewName := "message"
//
//	if v, err := g.SetView(viewName, x0, y0, maxX-x0, y0+2); err != nil {
//		if err != gocui.ErrUnknownView {
//			return err
//		}
//
//		fmt.Fprint(v, message)
//		if _, err := setCurrentViewOnTop(g, viewName); err != nil {
//			return err
//		}
//
//		close := func(g *gocui.Gui, v *gocui.View) error {
//			g.SetViewOnBottom(viewName)
//			//			if _, err := setCurrentViewOnTop(g, nextView); err != nil {
//			//				return err
//			//			}
//			return nil
//		}
//
//		if err := g.SetKeybinding(viewName, gocui.KeyEnter, gocui.ModNone, close); err != nil {
//			log.Panicln(err)
//		}
//	}
//
//	return nil
//}
