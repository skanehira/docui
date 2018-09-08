package panel

import (
	"fmt"
	"log"

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

	i.LoadImages(v)
	v.SetCursor(0, 1)

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
	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlC, gocui.ModNone, i.CreateContainerPanel); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlP, gocui.ModNone, i.PullImagePanel); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlD, gocui.ModNone, i.RemoveImage); err != nil {
		log.Panicln(err)
	}
}

func (i ImageList) CreateContainerPanel(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()

	id := i.GetImageID(v)
	if id == "" {
		return nil
	}

	data := map[string]interface{}{
		"Image": id,
	}

	input := NewInput(i.Gui, CreateContainerPanel, maxX/8, maxY/8, maxX-maxX/4-2, maxY-maxY/4-2, NewCreateContainerItems(), data)
	input.Init(i.Gui)
	return nil
}

func (i ImageList) PullImagePanel(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()
	input := NewInput(i.Gui, PullImagePanel, maxX/8, maxY/4, maxX-maxX/4-2, 8, NewPullImageItems(), make(map[string]interface{}))
	input.Init(i.Gui)
	return nil

}

func (i ImageList) DetailImage(g *gocui.Gui, v *gocui.View) error {

	id := i.GetImageID(v)
	if id == "" {
		return nil
	}

	img := i.Docker.InspectImage(id)

	nv, err := g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	nv.Clear()
	nv.SetOrigin(0, 0)
	fmt.Fprint(nv, StructToJson(img))

	return nil
}

func (i ImageList) RefreshPanel(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		nv, err := g.View(ImageListPanel)
		if err != nil {
			return err
		}

		v = nv
	}
	v.Clear()
	i.LoadImages(v)
	SetCurrentPanel(g, v.Name())
	return nil
}

func (i ImageList) LoadImages(v *gocui.View) {
	fmt.Fprintf(v, "%-15s %-20s\n", "ID", "NAME")
	for _, i := range i.Docker.Images() {
		fmt.Fprintf(v, "%-15s %-20s\n", i.ID[7:19], i.RepoTags)
	}
}

func (i ImageList) GetImageID(v *gocui.View) string {
	id := ReadLine(v, nil)
	if id == "" || id[:2] == "ID" {
		return ""
	}

	return id[:12]
}

func (i ImageList) RemoveImage(g *gocui.Gui, v *gocui.View) error {
	name := i.GetImageID(v)
	if name == "" {
		return nil
	}

	if err := i.Docker.RemoveImageWithName(name); err != nil {
		i.DispMessage(err.Error(), ImageListPanel)
		return nil
	}

	if err := i.RefreshPanel(g, v); err != nil {
		return err
	}

	return nil
}
