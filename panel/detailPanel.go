package panel

import (
	"log"

	"github.com/jroimartin/gocui"
)

type Detail struct {
	*Gui
	name string
	Position
}

func NewDetail(gui *Gui, name string, x, y, w, h int) Detail {
	return Detail{gui, name, Position{x, y, x + w, y + h}}
}

func (i Detail) Name() string {
	return i.name
}

func (i Detail) SetView(g *gocui.Gui) (*gocui.View, error) {
	v, err := g.SetView(i.Name(), i.x, i.y, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}

		v.Title = v.Name()
		v.Wrap = true
		return v, nil
	}

	return v, nil
}

func (i Detail) Init(g *Gui) {
	_, err := i.SetView(g.Gui)

	if err != nil {
		panic(err)
	}

	// keybinds
	g.SetKeybinds(i.Name())

	if err := g.SetKeybinding(i.Name(), Key("j"), gocui.ModNone, CursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("k"), gocui.ModNone, CursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("d"), gocui.ModNone, PageDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("u"), gocui.ModNone, PageUp); err != nil {
		log.Panicln(err)
	}
}

func (i Detail) RefreshPanel(g *gocui.Gui, v *gocui.View) error {
	//	v.Clear()
	//	SetCurrentPanel(g, v.Name())
	return nil
}
