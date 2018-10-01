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
	return Detail{gui, name, Position{x, y, w, h}}
}

func (i Detail) Name() string {
	return i.name
}

func (i Detail) SetView(g *gocui.Gui) error {
	v, err := g.SetView(i.Name(), i.x, i.y, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = v.Name()
		v.Wrap = true
	}

	i.SetKeyBinding()

	return nil
}

func (d Detail) SetKeyBinding() {
	d.SetKeyBindingToPanel(d.name)

	if err := d.SetKeybinding(d.name, 'j', gocui.ModNone, CursorDown); err != nil {
		log.Panicln(err)
	}
	if err := d.SetKeybinding(d.name, 'k', gocui.ModNone, CursorUp); err != nil {
		log.Panicln(err)
	}
	if err := d.SetKeybinding(d.name, 'd', gocui.ModNone, PageDown); err != nil {
		log.Panicln(err)
	}
	if err := d.SetKeybinding(d.name, 'u', gocui.ModNone, PageUp); err != nil {
		log.Panicln(err)
	}
}

func (i Detail) Refresh(g *gocui.Gui, v *gocui.View) error {
	return nil
}
