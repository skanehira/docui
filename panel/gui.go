package panel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"docui/docker"

	"github.com/jroimartin/gocui"
)

var active = 0

const (
	ImageListPanel       = "image list"
	PullImagePanel       = "pull image"
	ContainerListPanel   = "container list"
	DetailPanel          = "detail"
	CreateContainerPanel = "create container"
	MessagePanel         = "message"
)

type Gui struct {
	*gocui.Gui
	Docker     *docker.Docker
	Panels     map[string]Panel
	PanelNames []string
	PrePanel   string
}

type Panel interface {
	Init(*Gui)
	SetView(*gocui.Gui) (*gocui.View, error)
	Name() string
	RefreshPanel(*gocui.Gui, *gocui.View) error
}

type Position struct {
	x, y int
	w, h int
}

func New(mode gocui.OutputMode) *Gui {
	g, err := gocui.NewGui(mode)
	if err != nil {
		panic(err)
	}

	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen

	d := docker.NewDocker()

	gui := &Gui{
		g,
		d,
		make(map[string]Panel),
		[]string{},
		"",
	}

	gui.init()

	return gui
}

func SetCurrentPanel(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func (g *Gui) AddPanels(panel Panel) {
	name := panel.Name()
	//if name == DetailPanel {
	//	return
	//}
	g.PanelNames = append(g.PanelNames, name)
}

func (g *Gui) SetKeybinds(panel string) {
	if err := g.SetKeybinding(panel, gocui.KeyCtrlQ, gocui.ModNone, g.quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding(panel, gocui.KeyTab, gocui.ModNone, g.nextPanel); err != nil {
		log.Panicln(err)
	}
}

func (gui *Gui) nextPanel(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(gui.PanelNames)
	name := gui.PanelNames[nextIndex]

	if _, err := SetCurrentPanel(g, name); err != nil {
		return err
	}

	active = nextIndex
	return nil
}

func (gui *Gui) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func CursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}

	return nil
}

func CursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}

	return nil
}

func PageDown(g *gocui.Gui, v *gocui.View) error {
	_, maxY := g.Size()
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+maxY/2); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+maxY/2); err != nil {
				return err
			}
		}
	}

	return nil
}

func PageUp(g *gocui.Gui, v *gocui.View) error {
	_, maxY := g.Size()
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-maxY/2); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-maxY/2); err != nil {
				return err
			}
		}
	}

	return nil
}

func ReadLine(v *gocui.View, y *int) string {
	if y == nil {
		_, ny := v.Cursor()
		y = &ny
	}

	str, err := v.Line(*y)

	if err != nil {
		return ""
	}

	return str
}

func (g *Gui) init() {
	maxX, maxY := g.Size()

	g.StorePanels(NewImageList(g, ImageListPanel, 0, 0, maxX/3, maxY/2))
	g.StorePanels(NewContainerList(g, ContainerListPanel, 0, maxY/2+1, maxX/3, maxY-(maxY/2)-2))
	g.StorePanels(NewDetail(g, DetailPanel, maxX/3+2, 0, maxX-(maxX/3)-3, maxY-1))
}

func (g *Gui) StorePanels(panel Panel) {
	g.Panels[panel.Name()] = panel
	panel.Init(g)
	g.AddPanels(panel)
}

func (gui *Gui) DispMessage(message string, prePanel string) {
	gui.PrePanel = prePanel
	maxX, maxY := gui.Size()
	x := maxX / 5
	y := maxY / 3
	v, err := gui.SetView(MessagePanel, x, y, maxX-x, y+4)
	if err != nil {
		if err != gocui.ErrUnknownView {
			panic(err)
		}
		v.Wrap = true
		v.Title = MessagePanel
		fmt.Fprint(v, message)
		SetCurrentPanel(gui.Gui, v.Name())
	}

	if err := gui.SetKeybinding(v.Name(), gocui.KeyEnter, gocui.ModNone, gui.CloseMessage); err != nil {
		panic(err)
	}
}

func (gui *Gui) CloseMessage(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(v.Name()); err != nil {
		panic(err)
	}

	g.DeleteKeybindings(v.Name())
	gui.RefreshAllPanel()
	return nil
}

func (gui *Gui) RefreshAllPanel() {
	for _, panel := range gui.Panels {
		panel.RefreshPanel(gui.Gui, nil)
	}

	SetCurrentPanel(gui.Gui, gui.PrePanel)
}

func StructToJson(i interface{}) string {
	j, err := json.Marshal(i)
	if err != nil {
		return ""
	}

	out := new(bytes.Buffer)
	json.Indent(out, j, "", "    ")
	return out.String()
}
