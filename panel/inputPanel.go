package panel

import (
	"fmt"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

var activeInput = 0

type Handlers map[interface{}]func(g *gocui.Gui, v *gocui.View) error

type Input struct {
	*Gui
	name string
	Position
	Items
	Data     map[string]interface{}
	Handlers Handlers
}

type Item struct {
	Label map[string]Position
	Input map[string]Position
}

type Items []Item

func NewInput(g *Gui, name string, x, y, w, h int, items Items, data map[string]interface{}, handlers Handlers) Input {
	i := Input{
		Gui:      g,
		name:     name,
		Position: Position{x, y, w, h},
		Items:    items,
		Data:     data,
		Handlers: handlers,
	}

	g.SetNaviWithPanelName(name)

	g.StorePanels(i)

	if err := i.SetView(g.Gui); err != nil {
		panic(err)
	}

	return i
}

func (i Input) Name() string {
	return i.name
}

func (i Input) SetView(g *gocui.Gui) error {

	v, err := g.SetView(i.Name(), i.x, i.y, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = v.Name()
		v.Autoscroll = true
		v.Wrap = true
	}

	// create input panels
	for index, item := range i.Items {
		for name, p := range item.Label {
			if v, err := g.SetView(name, i.x+p.x, i.y+p.y, i.x+p.w, i.y+p.h); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				v.Wrap = true
				v.Frame = false
				fmt.Fprint(v, name)
			}
		}

		for name, p := range item.Input {
			if v, err := g.SetView(name, i.x+p.x, i.y+p.y, i.x+p.w, i.y+p.h); err != nil {
				if err != gocui.ErrUnknownView {
					return err
				}
				v.Wrap = true
				v.Editable = true
				v.Editor = i

				if index == 0 {
					SetCurrentPanel(g, name)
				}

				if name == "ImageInput" {
					fmt.Fprint(v, i.Data["Image"])
				}

				if name == "ContainerInput" {
					fmt.Fprint(v, i.Data["Container"])
				}

				// set kyebinding
				for k, handler := range i.Handlers {
					if err := i.SetKeybinding(name, k, gocui.ModNone, handler); err != nil {
						panic(err)
					}
				}

				i.SetKeyBinding(name)
			}
		}
	}

	return nil
}

func (i Input) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(+1, 0, false)
	}
}

func (i Input) SetKeyBinding(name string) {
	if err := i.SetKeybinding(name, gocui.KeyCtrlJ, gocui.ModNone, i.NextItem); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(name, gocui.KeyCtrlK, gocui.ModNone, i.PreItem); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(name, gocui.KeyCtrlW, gocui.ModNone, i.ClosePanel); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(name, gocui.KeyEsc, gocui.ModNone, i.ClosePanel); err != nil {
		panic(err)
	}
}

func (i Input) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	activeInput = 0

	i.CloseItemPanel()

	if err := i.DeleteView(i.Name()); err != nil {
		panic(err)
	}
	i.DeleteKeybindings(i.Name())

	if i.NextPanel == "" {
		i.NextPanel = ImageListPanel
	}

	i.SwitchPanel(i.NextPanel)

	return nil
}

func (i Input) CloseItemPanel() {
	for _, item := range i.Items {
		if err := i.DeleteView(i.GetKeyFromMap(item.Label)); err != nil {
			panic(err)
		}

		name := i.GetKeyFromMap(item.Input)
		i.DeleteKeybindings(name)

		if err := i.DeleteView(name); err != nil {
			panic(err)
		}
	}
}

func (i Input) NextItem(g *gocui.Gui, v *gocui.View) error {

	nextIndex := (activeInput + 1) % len(i.Items)
	item := i.Items[nextIndex]

	name := i.GetKeyFromMap(item.Input)

	if _, err := SetCurrentPanel(g, name); err != nil {
		return err
	}

	activeInput = nextIndex
	return nil
}

func (i Input) PreItem(g *gocui.Gui, v *gocui.View) error {
	nextIndex := activeInput - 1
	if nextIndex < 0 {
		nextIndex = len(i.Items) - 1
	} else {
		nextIndex = (activeInput - 1) % len(i.Items)
	}

	item := i.Items[nextIndex]

	name := i.GetKeyFromMap(item.Input)

	if _, err := SetCurrentPanel(g, name); err != nil {
		return err
	}

	activeInput = nextIndex
	return nil
}

func (i Input) Refresh() error {
	i.Update(func(g *gocui.Gui) error {
		return nil
	})

	return nil
}

func NewItems(labels []string, ix, iy, iw, ih, wl int) Items {

	var items Items

	x := iw / 8                                            // label start position
	w := x + wl                                            // label length
	bh := 2                                                // input box height
	th := ((ih - iy) - len(labels)*bh) / (len(labels) + 1) // to next input height
	y := th
	h := 0

	for i, name := range labels {
		if i != 0 {
			y = items[i-1].Label[labels[i-1]].h + th
		}
		h = y + bh

		x1 := w + 1
		w1 := iw - (x + ix)

		item := Item{
			Label: map[string]Position{name: {x, y, w, h}},
			Input: map[string]Position{name + "Input": {x1, y, w1, h}},
		}

		items = append(items, item)
	}

	return items
}

func ParseDateToString(unixtime int64) string {
	t := time.Unix(unixtime, 0)
	return t.Format("2006/01/02 15:04:05")
}

func ParseSizeToString(size int64) string {
	mb := float64(size) / 1024 / 1024
	return fmt.Sprintf("%.1fMB", mb)
}

func ParsePortToString(ports []docker.APIPort) string {
	var port string
	for _, p := range ports {
		if p.PublicPort == 0 {
			port += fmt.Sprintf("%d/%s ", p.PrivatePort, p.Type)
		} else {
			port += fmt.Sprintf("%s:%d->%d/%s ", p.IP, p.PublicPort, p.PrivatePort, p.Type)
		}
	}
	return port
}
