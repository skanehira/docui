package panel

import (
	"fmt"
	"log"
	"strings"

	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"

	"github.com/jroimartin/gocui"
)

const (
	ImageListPanel               = "image list scroll"
	ImageListHeaderPanel         = "image list"
	PullImagePanel               = "pull image"
	ContainerListPanel           = "container list scroll"
	ContainerListHeaderPanel     = "container list"
	DetailPanel                  = "detail"
	CreateContainerPanel         = "create container"
	ErrMessagePanel              = "error message"
	SaveImagePanel               = "save image"
	ImportImagePanel             = "import image"
	LoadImagePanel               = "load image"
	ExportContainerPanel         = "export container"
	CommitContainerPanel         = "commit container"
	RenameContainerPanel         = "rename container"
	ConfirmMessagePanel          = "confirm"
	StateMessagePanel            = "state"
	SearchImagePanel             = "search images"
	SearchImageResultPanel       = "images scroll"
	SearchImageResultHeaderPanel = "images"
	VolumeListPanel              = "volume list scroll"
	VolumeListHeaderPanel        = "volume list"
	CreateVolumePanel            = "create volume"
	NavigatePanel                = "navigate"
)

type Gui struct {
	*gocui.Gui
	Docker     *docker.Docker
	Panels     map[string]Panel
	PanelNames []string
	NextPanel  string
	active     int
}

type Panel interface {
	SetView(*gocui.Gui) error
	Name() string
	Refresh(*gocui.Gui, *gocui.View) error
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
	g.SelFgColor = gocui.AttrBold
	g.InputEsc = true

	d := docker.NewDocker()

	gui := &Gui{
		Gui:        g,
		Docker:     d,
		Panels:     make(map[string]Panel),
		PanelNames: []string{},
		NextPanel:  ImageListPanel,
		active:     0,
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
	if err := g.SetKeybinding(panel, gocui.KeyCtrlD, gocui.ModNone, g.DockerInfo); err != nil {
		log.Panicln(err)
	}
}

func (g *Gui) SetGlobalKeyBinding() {
	if err := g.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, g.quit); err != nil {
		log.Panicln(err)
	}
}

func (gui *Gui) DockerInfo(g *gocui.Gui, v *gocui.View) error {
	gui.PopupDetailPanel(g, v)

	v, err := g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	if info, err := gui.Docker.Info(); err == nil {
		fmt.Fprint(v, common.StructToJson(info))
	}

	return nil
}

func (gui *Gui) nextPanel(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (gui.active + 1) % len(gui.PanelNames)
	name := gui.PanelNames[nextIndex]

	gui.SwitchPanel(name)
	gui.active = nextIndex
	return nil
}

func (gui *Gui) prePanel(g *gocui.Gui, v *gocui.View) error {
	nextIndex := gui.active - 1

	if nextIndex < 0 {
		nextIndex = len(gui.PanelNames) - 1
	} else {
		nextIndex = (gui.active - 1) % len(gui.PanelNames)
	}

	name := gui.PanelNames[nextIndex]
	gui.SwitchPanel(name)

	gui.active = nextIndex

	return nil
}

func (gui *Gui) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (g *Gui) init() {
	maxX, maxY := g.Size()

	g.StorePanels(NewImageList(g, ImageListPanel, 0, 0, maxX-1, maxY/3-1))
	g.StorePanels(NewContainerList(g, ContainerListPanel, 0, maxY/3, maxX-1, (maxY/3)*2-1))
	g.StorePanels(NewVolumeList(g, VolumeListPanel, 0, maxY/3*2, maxX-1, maxY-3))
	g.StorePanels(NewNavigate(g, NavigatePanel, 0, maxY-3, maxX-1, maxY))

	for _, panel := range g.Panels {
		panel.SetView(g.Gui)
	}

	g.SwitchPanel(ImageListPanel)
	g.SetGlobalKeyBinding()
}

func (g *Gui) StorePanels(panel Panel) {
	g.Panels[panel.Name()] = panel

	storeTarget := map[string]bool{
		ImageListPanel:     true,
		ContainerListPanel: true,
		DetailPanel:        true,
		VolumeListPanel:    true,
	}

	if storeTarget[panel.Name()] {
		g.AddPanelNames(panel)
	}

}

func (gui *Gui) PopupDetailPanel(g *gocui.Gui, v *gocui.View) error {
	gui.NextPanel = g.CurrentView().Name()

	maxX, maxY := g.Size()
	panel := NewDetail(gui, DetailPanel, maxX/7, 1, maxX-(maxX/7), maxY-4)

	panel.SetView(g)

	return nil
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
		v, _ := gui.View(panel.Name())
		panel.Refresh(gui.Gui, v)
	}

	gui.SwitchPanel(gui.NextPanel)
}

func (gui *Gui) SwitchPanel(next string) *gocui.View {
	v := gui.CurrentView()
	if v != nil {
		v.Highlight = false
	}

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
	return navi.SetNavigate(name)
}

func (g *Gui) GetKeyFromMap(m map[string]Position) string {
	var key string
	for k, _ := range m {
		key = k
	}

	return key
}

func (g *Gui) GetItemsToMap(items Items) (map[string]string, error) {

	data := make(map[string]string)

	for _, item := range items {
		name := g.GetKeyFromMap(item.Label)

		v, err := g.View(g.GetKeyFromMap(item.Input))

		if err != nil {
			return data, err
		}

		value := ReadLine(v, nil)

		if value == "" {
			if name == "Tag" {
				value = "latest"
			}
		}

		data[name] = value
	}

	return data, nil
}

func SetCurrentPanel(g *gocui.Gui, name string) (*gocui.View, error) {
	v, err := g.SetCurrentView(name)

	if err != nil {
		return nil, err
	}

	v.Highlight = true

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
