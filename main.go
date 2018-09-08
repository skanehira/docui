package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

var (
	viewArr = []string{"v1", "v2", "v3"}
	active  = 0
)

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
	if v, err := g.SetView("v1", 0, 0, maxX/2-1, maxY/2-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "image list"
		v.Autoscroll = true
		v.Wrap = true

		if _, err = setCurrentViewOnTop(g, "v1"); err != nil {
			return err
		}

		for _, i := range d.Images() {
			fmt.Fprintf(v, "%+v\n", i.RepoTags)
		}
	}

	if v, err := g.SetView("v2", 0, maxY/2-1, maxX/2-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "container list"
		v.Wrap = true
		v.Autoscroll = true

		for _, c := range d.Containers() {
			fmt.Fprintf(v, "name:%+v status:%s\n", c.Names, c.Status)
		}
	}
	if v, err := g.SetView("v3", maxX/2-1, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "detail"
		v.Wrap = true
		v.Autoscroll = true
	}
	//	if v, err := g.SetView("v4", maxX/2, maxY/2, maxX-1, maxY-1); err != nil {
	//		if err != gocui.ErrUnknownView {
	//			return err
	//		}
	//		v.Title = "v4 (editable)"
	//		v.Editable = true
	//	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

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

	// select image
	if err := g.SetKeybinding("v1", gocui.KeyCtrlJ, gocui.ModNone, d.NextImage); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("v1", gocui.KeyCtrlK, gocui.ModNone, d.PreImage); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("v1", gocui.KeyEnter, gocui.ModNone, d.DetailImage); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func (d *Docker) NextImage(g *gocui.Gui, v *gocui.View) error {
	g.DeleteView("image_detail")

	x, y := v.Cursor()
	nextY := y + 1
	str, _ := v.Line(nextY)

	if str != "" {
		v.SetCursor(x, nextY)
	}

	return nil
}

func (d *Docker) PreImage(g *gocui.Gui, v *gocui.View) error {
	g.DeleteView("image_detail")

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

	nv, err := g.View("v3")
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
		//"Config":          true,
		"ContainerConfig": true,
	}

	for i := 0; i < t.NumField(); i++ {
		// field name
		fieldName := t.Field(i).Name

		// none display
		if noDisplay[fieldName] {
			continue
		}

		fmt.Fprintf(nv, "%-15s: ", fieldName)

		value := value.Field(i)

		if fieldName == "RepoTags" {
			for _, v := range value.Interface().([]string) {
				fmt.Fprintf(nv, "%s\n", v)
			}
			continue
		}

		if fieldName == "Created" {
			fmt.Fprintf(nv, "%s\n", value.Interface().(time.Time))
			continue
		}

		if fieldName == "Size" || fieldName == "VirtualSize" {
			fmt.Fprintf(nv, "%d\n", value.Int())
			continue
		}

		if fieldName == "Config" {
			c := value.Interface().(*docker.Config)

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

		//	if fieldName == "ContainerConfig" {
		//		fmt.Fprintf(nv, "%+v\n", value.Interface().(docker.Config))
		//		continue
		//	}

		fmt.Fprintf(nv, "%s\n", value.String())
	}

	return nil
}
