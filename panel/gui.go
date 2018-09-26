package panel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/skanehira/docui/docker"

	"github.com/jroimartin/gocui"
)

var active = 0

const (
	ImageListPanel         = "image list"
	PullImagePanel         = "pull image"
	ContainerListPanel     = "container list"
	DetailPanel            = "detail"
	CreateContainerPanel   = "create container"
	ErrMessagePanel        = "error message"
	SaveImagePanel         = "save image"
	ImportImagePanel       = "import image"
	LoadImagePanel         = "load image"
	ExportContainerPanel   = "export container"
	CommitContainerPanel   = "commit container"
	ConfirmMessagePanel    = "confirm"
	StateMessagePanel      = "state"
	SearchImagePanel       = "search images"
	SearchImageResultPanel = "images"
	NavigatePanel          = "navigate"
)

type Gui struct {
	*gocui.Gui
	Docker     *docker.Docker
	Panels     map[string]Panel
	PanelNames []string
	NextPanel  string
}

type Panel interface {
	SetView(*gocui.Gui) error
	Name() string
	Refresh() error
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
	g.InputEsc = true

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

func (g *Gui) AddPanelNames(panel Panel) {
	name := panel.Name()
	g.PanelNames = append(g.PanelNames, name)
}

func (g *Gui) SetKeyBindingToPanel(panel string) {
	if err := g.SetKeybinding(panel, 'q', gocui.ModNone, g.quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(panel, 'h', gocui.ModNone, g.prePanel); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(panel, 'l', gocui.ModNone, g.nextPanel); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(panel, gocui.KeyTab, gocui.ModNone, g.nextPanel); err != nil {
		log.Panicln(err)
	}
}

func (g *Gui) SetGlobalKeyBinding() {
	if err := g.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, g.quit); err != nil {
		log.Panicln(err)
	}
}

func (gui *Gui) nextPanel(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(gui.PanelNames)
	name := gui.PanelNames[nextIndex]

	gui.SwitchPanel(name)
	active = nextIndex
	return nil
}

func (gui *Gui) prePanel(g *gocui.Gui, v *gocui.View) error {
	nextIndex := active - 1

	if nextIndex < 0 {
		nextIndex = len(gui.PanelNames) - 1
	} else {
		nextIndex = (active - 1) % len(gui.PanelNames)
	}

	name := gui.PanelNames[nextIndex]
	gui.SwitchPanel(name)

	active = nextIndex

	return nil
}

func (gui *Gui) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (g *Gui) init() {
	maxX, maxY := g.Size()

	g.StorePanels(NewImageList(g, ImageListPanel, 0, 0, maxX/2, maxY/2))
	g.StorePanels(NewContainerList(g, ContainerListPanel, 0, maxY/2+1, maxX/2, maxY-(maxY/2)-4))
	g.StorePanels(NewDetail(g, DetailPanel, maxX/2+2, 0, maxX-(maxX/2)-3, maxY-3))
	g.StorePanels(NewNavigate(g, NavigatePanel, 0, maxY-3, maxX-1, 5))

	for _, panel := range g.Panels {
		panel.SetView(g.Gui)
	}

	g.SwitchPanel(ImageListPanel)
	g.SetGlobalKeyBinding()

	//monitoring container status interval 5s
	go func() {
		c := g.Panels[ContainerListPanel].(ContainerList)

		for {
			c.Update(func(g *gocui.Gui) error {
				c.Refresh()
				return nil
			})
			time.Sleep(5 * time.Second)
		}
	}()
}

func (g *Gui) StorePanels(panel Panel) {
	g.Panels[panel.Name()] = panel

	if panel.Name() != NavigatePanel {
		g.AddPanelNames(panel)
	}
}

func (gui *Gui) ErrMessage(message string, nextPanel string) {
	gui.Update(func(g *gocui.Gui) error {
		gui.NextPanel = nextPanel
		maxX, maxY := gui.Size()

		x := maxX / 5
		y := maxY / 3
		v, err := gui.SetView(ErrMessagePanel, x, y, maxX-x, y+4)
		if err != nil {
			if err != gocui.ErrUnknownView {
				panic(err)
			}
			v.Wrap = true
			v.Title = v.Name()
			fmt.Fprint(v, message)
			gui.SwitchPanel(v.Name())
		}

		if err := gui.SetKeybinding(v.Name(), gocui.KeyEnter, gocui.ModNone, gui.CloseMessage); err != nil {
			panic(err)
		}

		return nil
	})
}

func (gui *Gui) CloseMessage(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(v.Name()); err != nil {
		panic(err)
	}
	g.DeleteKeybindings(v.Name())
	gui.RefreshAllPanel()
	return nil
}

func (gui *Gui) ConfirmMessage(message string, f func(g *gocui.Gui, v *gocui.View) error) {
	maxX, maxY := gui.Size()
	x := maxX / 5
	y := maxY / 3
	v, err := gui.SetView(ConfirmMessagePanel, x, y, maxX-x, y+2)
	if err != nil {
		if err != gocui.ErrUnknownView {
			panic(err)
		}
		v.Wrap = true
		v.Title = v.Name()
		fmt.Fprint(v, message)
		gui.SwitchPanel(v.Name())
	}

	if err := gui.SetKeybinding(v.Name(), 'y', gocui.ModNone, f); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(v.Name(), gocui.KeyEnter, gocui.ModNone, f); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(v.Name(), 'n', gocui.ModNone, gui.CloseConfirmMessage); err != nil {
		panic(err)
	}
}

func (gui *Gui) CloseConfirmMessage(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(ConfirmMessagePanel); err != nil {
		panic(err)
	}

	g.DeleteKeybindings(ConfirmMessagePanel)
	gui.SwitchPanel(gui.NextPanel)
	return nil
}

func (gui *Gui) StateMessage(message string) *gocui.View {
	maxX, maxY := gui.Size()
	x := maxX / 3
	y := maxY / 3
	v, err := gui.SetView(StateMessagePanel, x, y, maxX-x, y+2)
	if err != nil {
		if err != gocui.ErrUnknownView {
			panic(err)
		}
		v.Wrap = true
		v.Title = v.Name()
		fmt.Fprint(v, message)

		gui.SwitchPanel(v.Name())
	}

	return v
}

func (gui *Gui) CloseStateMessage() {
	if err := gui.DeleteView(StateMessagePanel); err != nil {
		panic(err)
	}
}

func (gui *Gui) RefreshAllPanel() {
	for _, panel := range gui.Panels {
		panel.Refresh()
	}

	gui.SwitchPanel(gui.NextPanel)
}

func (gui *Gui) SwitchPanel(next string) *gocui.View {
	v, err := SetCurrentPanel(gui.Gui, next)
	if err != nil {
		panic(err)
	}

	gui.SetNaviWithPanelName(next)
	return v
}

func (g *Gui) IsSetView(name string) bool {
	if v, err := g.View(name); err != nil && v == nil {
		return false
	}

	return true
}

func (g *Gui) SetNaviWithPanelName(name string) *gocui.View {
	navi := g.Panels[NavigatePanel].(Navigate)
	return navi.SetNavi(name)
}

func SetCurrentPanel(g *gocui.Gui, name string) (*gocui.View, error) {
	_, err := g.SetCurrentView(name)
	if err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func CursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		nexty := cy + 1

		line := ReadLine(v, &nexty)
		if line == "" {
			return nil
		}

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

		if (v.Name() == ImageListPanel || v.Name() == ContainerListPanel || v.Name() == SearchImageResultPanel) && cy-1 == 0 && oy-1 < 1 {
			return nil
		}

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

	return strings.Trim(str, " ")
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
