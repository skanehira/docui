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

	if err := g.SetKeybinding(i.Name(), Key("j"), gocui.ModNone, CursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("k"), gocui.ModNone, CursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyEnter, gocui.ModNone, i.DetailImage); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("c"), gocui.ModNone, i.CreateContainerPanel); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("p"), gocui.ModNone, i.PullImagePanel); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("d"), gocui.ModNone, i.RemoveImage); err != nil {
		log.Panicln(err)
	}
}

func (i ImageList) CreateContainerPanel(g *gocui.Gui, v *gocui.View) error {
	id := i.GetImageID(v)
	if id == "" {
		return nil
	}

	data := map[string]interface{}{
		"Image": id,
	}

	maxX, maxY := i.Size()
	x := maxX / 8
	y := maxY / 8
	w := maxX - x
	h := maxY - y
	input := NewInput(i.Gui, CreateContainerPanel, x, y, w, h, NewCreateContainerItems(x, y, w, h), data)
	input.Init(i.Gui)
	return nil
}

func (i ImageList) PullImagePanel(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	input := NewInput(i.Gui, PullImagePanel, x, y, w, h, NewPullImageItems(x, y, w, h), make(map[string]interface{}))
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
