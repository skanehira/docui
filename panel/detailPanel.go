package panel

import (
	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

// Detail panel
type Detail struct {
	*Gui
	name string
	Position
}

// NewDetail create new detail panel.
func NewDetail(gui *Gui, name string, x, y, w, h int) Detail {
	return Detail{gui, name, Position{x, y, w, h}}
}

// Name return panel name.
func (d Detail) Name() string {
	return d.name
}

// SetView set up detail panel.
func (d Detail) SetView(g *gocui.Gui) error {
	v, err := common.SetViewWithValidPanelSize(g, d.Name(), d.x, d.y, d.w, d.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = v.Name()
		v.Wrap = true
		v.SelBgColor = gocui.ColorCyan
	}

	d.SetKeyBinding()
	d.SwitchPanel(d.name)

	return nil
}

// SetKeyBinding set key bind to this panel.
func (d Detail) SetKeyBinding() {

	if err := d.SetKeybinding(d.name, 'j', gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := d.SetKeybinding(d.name, 'k', gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := d.SetKeybinding(d.name, 'd', gocui.ModNone, PageDown); err != nil {
		panic(err)
	}
	if err := d.SetKeybinding(d.name, 'u', gocui.ModNone, PageUp); err != nil {
		panic(err)
	}
	if err := d.SetKeybinding(d.name, gocui.KeyEsc, gocui.ModNone, d.CloseDetailPanel); err != nil {
		panic(err)
	}
	if err := d.SetKeybinding(d.name, 'q', gocui.ModNone, d.CloseDetailPanel); err != nil {
		panic(err)
	}
	if err := d.SetKeybinding(d.name, gocui.KeyCtrlQ, gocui.ModNone, d.quit); err != nil {
		panic(err)
	}
}

// Refresh do nothing
func (d Detail) Refresh(g *gocui.Gui, v *gocui.View) error {
	return nil
}

// CloseDetailPanel close the detail panel.
func (d Detail) CloseDetailPanel(g *gocui.Gui, v *gocui.View) error {

	if err := d.DeleteView(d.Name()); err != nil {
		common.Logger.Error(err)
		return err
	}
	d.DeleteKeybindings(d.Name())

	if d.NextPanel == "" {
		d.NextPanel = ImageListPanel
	}

	d.SwitchPanel(d.NextPanel)

	return nil
}
