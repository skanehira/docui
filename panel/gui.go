package panel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/skanehira/docui/docker"

	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

var active = 0

const (
	ImageListPanel       = "image list"
	PullImagePanel       = "pull image"
	ContainerListPanel   = "container list"
	DetailPanel          = "detail"
	CreateContainerPanel = "create container"
	ErrMessagePanel      = "error message"
	SaveImagePanel       = "save image"
	ImportImagePanel     = "import image"
	LoadImagePanel       = "load image"
	ExportContainerPanel = "export container"
	CommitContainerPanel = "commit container"
	ConfirmMessagePanel  = "confirm"
	StateMessagePanel    = "state"
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
	_, err := g.SetCurrentView(name)
	if err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func (g *Gui) AddPanelNames(panel Panel) {
	name := panel.Name()
	g.PanelNames = append(g.PanelNames, name)
}

func (g *Gui) SetKeybinds(panel string) {
	if err := g.SetKeybinding(panel, gocui.KeyCtrlQ, gocui.ModNone, g.quit); err != nil {
		log.Panicln(err)
	}
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

func (gui *Gui) nextPanel(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(gui.PanelNames)
	name := gui.PanelNames[nextIndex]

	if _, err := SetCurrentPanel(g, name); err != nil {
		return err
	}

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

	g.StorePanels(NewImageList(g, ImageListPanel, 0, 0, maxX/2, maxY/2))
	g.StorePanels(NewContainerList(g, ContainerListPanel, 0, maxY/2+1, maxX/2, maxY-(maxY/2)-2))
	g.StorePanels(NewDetail(g, DetailPanel, maxX/2+2, 0, maxX-(maxX/2)-3, maxY-1))

	for _, panel := range g.Panels {
		panel.SetView(g.Gui)
	}

	if _, err := SetCurrentPanel(g.Gui, ImageListPanel); err != nil {
		panic(err)
	}

	// monitoring container status interval 5s
	go func() {
		c := g.Panels[ContainerListPanel].(ContainerList)
		v, err := g.View(ContainerListPanel)
		if err != nil {
			panic(err)
		}

		for {
			c.Update(func(g *gocui.Gui) error {
				c.GetContainerList(v)
				return nil
			})
			time.Sleep(5 * time.Second)
		}
	}()
}

func (g *Gui) StorePanels(panel Panel) {
	g.Panels[panel.Name()] = panel
	g.AddPanelNames(panel)
}

func (gui *Gui) ErrMessage(message string, nextPanel string) {
	gui.Update(func(g *gocui.Gui) error {
		gui.NextPanel = nextPanel
		func() {
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
				SetCurrentPanel(gui.Gui, v.Name())
			}

			if err := gui.SetKeybinding(v.Name(), gocui.KeyEnter, gocui.ModNone, gui.CloseMessage); err != nil {
				panic(err)
			}
		}()

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
		v.Title = ConfirmMessagePanel
		fmt.Fprint(v, message)
		SetCurrentPanel(gui.Gui, v.Name())
	}

	if err := gui.SetKeybinding(v.Name(), 'y', gocui.ModNone, f); err != nil {
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
	SetCurrentPanel(gui.Gui, gui.NextPanel)
	return nil
}

func (gui *Gui) StateMessage(message string) *gocui.View {
	maxX, maxY := gui.Size()
	x := maxX / 5
	y := maxY / 3
	v, err := gui.SetView(StateMessagePanel, x, y, maxX-x, y+2)
	if err != nil {
		if err != gocui.ErrUnknownView {
			panic(err)
		}
		v.Wrap = true
		v.Title = v.Name()
		fmt.Fprint(v, message)
		if _, err := SetCurrentPanel(gui.Gui, v.Name()); err != nil {
			panic(err)
		}
	}

	return v
}

func (gui *Gui) CloseStateMessage(v *gocui.View) {
	if err := gui.DeleteView(v.Name()); err != nil {
		panic(err)
	}
}

func (gui *Gui) RefreshAllPanel() {
	for _, panel := range gui.Panels {
		panel.Refresh()
	}

	SetCurrentPanel(gui.Gui, gui.NextPanel)
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

func ParseDateToString(unixtime int64) string {
	t := time.Unix(unixtime, 0)
	return t.Format("2006/01/02 15:04:05")
}

func ParseSizeToString(size int64) string {
	mb := float64(size) / 1024 / 1024
	return fmt.Sprintf("%.1fMB", mb)
}

func ParsePortToString(ports []dockerclient.APIPort) string {
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
