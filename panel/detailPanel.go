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

func (d Detail) Name() string {
	return d.name
}

func (d Detail) SetView(g *gocui.Gui) error {
	v, err := g.SetView(d.Name(), d.x, d.y, d.w, d.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = v.Name()
		v.Wrap = true
	}

	d.SetKeyBinding()
	d.SwitchPanel(d.name)

	return nil
}

func (d Detail) SetKeyBinding() {

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
	if err := d.SetKeybinding(d.name, gocui.KeyEsc, gocui.ModNone, d.CloseDetailPanel); err != nil {
		log.Panicln(err)
	}
	if err := d.SetKeybinding(d.name, 'q', gocui.ModNone, d.CloseDetailPanel); err != nil {
		log.Panicln(err)
	}
	if err := d.SetKeybinding(d.name, gocui.KeyCtrlQ, gocui.ModNone, d.quit); err != nil {
		log.Panicln(err)
	}
}

func (d Detail) Refresh(g *gocui.Gui, v *gocui.View) error {
	return nil
}

func (d Detail) CloseDetailPanel(g *gocui.Gui, v *gocui.View) error {

	if err := d.DeleteView(d.Name()); err != nil {
		panic(err)
	}
	d.DeleteKeybindings(d.Name())

	if d.NextPanel == "" {
		d.NextPanel = ImageListPanel
	}

	d.SwitchPanel(d.NextPanel)

	return nil
}
