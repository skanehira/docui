package panel

import (
	"errors"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
	component "github.com/skanehira/gocui-component"
)

const (
	ImageListPanel               = "image list scroll"
	ImageListHeaderPanel         = "image list"
	PullImagePanel               = "pull image"
	ContainerListPanel           = "container list scroll"
	ContainerListHeaderPanel     = "container list"
	DetailPanel                  = "detail"
	CreateContainerPanel         = "create container"
	SaveImagePanel               = "save image"
	ImportImagePanel             = "import image"
	LoadImagePanel               = "load image"
	ExportContainerPanel         = "export container"
	CommitContainerPanel         = "commit container"
	RenameContainerPanel         = "rename container"
	SearchImagePanel             = "search images"
	SearchImageResultPanel       = "images scroll"
	SearchImageResultHeaderPanel = "images"
	VolumeListPanel              = "volume list scroll"
	VolumeListHeaderPanel        = "volume list"
	CreateVolumePanel            = "create volume"
	NavigatePanel                = "navigate"
	InfoPanel                    = "info"
	DockerInfoPanel              = "docker info"
	HostInfoPanel                = "host info"
	FilterPanel                  = "filter"
	NetworkListPanel             = "network list scroll"
	NetworkListHeaderPanel       = "network list"
	TaskListHeaderPanel          = "task list"
	TaskListPanel                = "task list scroll"
	ExecContainerCmd             = "exec container cmd"
)

// ExecFlag use to attach container
// TODO imporvement this logic
var ExecFlag = errors.New("exec")

type Gui struct {
	*gocui.Gui
	Docker     *docker.Docker
	Panels     map[string]Panel
	PanelNames []string
	NextPanel  string
	active     int
	modal      *component.Modal
	Logger     *common.Logger
}

type Panel interface {
	SetView(*gocui.Gui) error
	Name() string
	Refresh(*gocui.Gui, *gocui.View) error
	Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier)
}

type Position struct {
	x, y int
	w, h int
}

func New(mode gocui.OutputMode, d *docker.Docker) *Gui {
	g, err := gocui.NewGui(mode)
	if err != nil {
		panic(err)
	}

	g.Highlight = true
	g.SelFgColor = gocui.AttrBold
	g.InputEsc = true

	gui := &Gui{
		Gui:        g,
		Docker:     d,
		Panels:     make(map[string]Panel),
		PanelNames: []string{},
		NextPanel:  ImageListPanel,
		active:     0,
		Logger:     common.NewLogger(),
	}

	gui.init()

	return gui
}

func (gui *Gui) Close() {
	gui.Gui.Close()
	gui.Logger.CloseLogger()
}

func (gui *Gui) AddPanelNames(panel Panel) {
	name := panel.Name()
	gui.PanelNames = append(gui.PanelNames, name)
}

func (gui *Gui) SetKeyBindingToPanel(panel string) {
	if err := gui.SetKeybinding(panel, 'q', gocui.ModNone, gui.quit); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(panel, 'h', gocui.ModNone, gui.prePanel); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(panel, gocui.KeyArrowLeft, gocui.ModNone, gui.prePanel); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(panel, 'l', gocui.ModNone, gui.nextPanel); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(panel, gocui.KeyArrowRight, gocui.ModNone, gui.nextPanel); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(panel, gocui.KeyTab, gocui.ModNone, gui.nextPanel); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(panel, 'j', gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(panel, gocui.KeyArrowDown, gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
    if err := gui.SetKeybinding(panel, gocui.KeyCtrlN, gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(panel, 'k', gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(panel, gocui.KeyArrowUp, gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
    if err := gui.SetKeybinding(panel, gocui.KeyCtrlP, gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
}

func (gui *Gui) SetGlobalKeyBinding() {
	if err := gui.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, gui.quit); err != nil {
		panic(err)
	}
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
	for _, task := range gui.Panels[TaskListPanel].(*TaskList).ViewTask {
		if task.Status == Executing.String() {
			gui.ErrMessage("task is running", gui.PanelNames[gui.active])
			return nil
		}
	}
	return gocui.ErrQuit
}

func (gui *Gui) init() {
	maxX, maxY := gui.Size()
	topY := maxY / 5

	gui.StorePanels(NewInfo(gui, DockerInfoPanel, 0, -1, maxX-1, 3))
	gui.StorePanels(NewTaskList(gui, TaskListPanel, 0, 3, maxX-1, topY-2))
	gui.StorePanels(NewImageList(gui, ImageListPanel, 0, topY-1, maxX-1, topY*2-2))
	gui.StorePanels(NewContainerList(gui, ContainerListPanel, 0, topY*2-1, maxX-1, topY*3-2))
	gui.StorePanels(NewVolumeList(gui, VolumeListPanel, 0, topY*3-1, maxX-1, topY*4-2))
	gui.StorePanels(NewNetworkList(gui, NetworkListPanel, 0, topY*4-1, maxX-1, maxY-3))
	gui.StorePanels(NewNavigate(gui, NavigatePanel, 0, maxY-3, maxX-1, maxY))

	for _, panel := range gui.Panels {
		if err := panel.SetView(gui.Gui); err != nil {
			panic(err)
		}
	}

	gui.SwitchPanel(ImageListPanel)
	gui.SetGlobalKeyBinding()
}

func (gui *Gui) StorePanels(panel Panel) {
	gui.Panels[panel.Name()] = panel

	storeTarget := map[string]bool{
		ImageListPanel:     true,
		ContainerListPanel: true,
		DetailPanel:        true,
		VolumeListPanel:    true,
		NetworkListPanel:   true,
		TaskListPanel:      true,
	}

	if storeTarget[panel.Name()] {
		gui.AddPanelNames(panel)
	}

}

func (gui *Gui) PopupDetailPanel(g *gocui.Gui, v *gocui.View) error {
	gui.NextPanel = g.CurrentView().Name()

	maxX, maxY := g.Size()
	panel := NewDetail(gui, DetailPanel, 0, 0, maxX-1, maxY-3)

	panel.SetView(g)

	return nil
}

func (gui *Gui) ErrMessage(message string, nextPanel string) {
	gui.Update(func(g *gocui.Gui) error {
		modal := gui.NewModal(message)

		cancelAction := func(g *gocui.Gui, v *gocui.View) error {
			modal.Close()
			gui.SwitchPanel(nextPanel)
			return nil
		}

		modal.AddButton("OK", gocui.KeyEnter, cancelAction).
			AddHandler(gocui.KeyEsc, cancelAction)

		modal.Draw()
		return nil
	})
}

func (gui *Gui) ConfirmMessage(message, next string, f func() error) {
	modal := gui.NewModal(message)

	cancelAction := func(g *gocui.Gui, v *gocui.View) error {
		modal.Close()
		gui.SwitchPanel(next)
		return nil
	}

	doAction := func(g *gocui.Gui, v *gocui.View) error {
		defer cancelAction(g, v)
		return f()
	}

	modal.AddButton("No", gocui.KeyEnter, cancelAction).
		AddHandler(gocui.KeyEsc, cancelAction).
		AddHandler('y', doAction).
		AddHandler('n', cancelAction)

	modal.AddButton("Yes", gocui.KeyEnter, doAction).
		AddHandler(gocui.KeyEsc, cancelAction).
		AddHandler('y', doAction).
		AddHandler('n', cancelAction)

	modal.Draw()
}

func (gui *Gui) StateMessage(message string) {
	gui.NewModal(message)
}

func (gui *Gui) CloseStateMessage() {
	gui.modal.Close()
}

func (gui *Gui) NewModal(message string) *component.Modal {
	maxX, maxY := gui.Size()
	x := maxX / 5
	y := maxY / 3
	w := x * 4

	modal := component.NewModal(gui.Gui, x, y, w).
		SetText(message)

	modal.Draw()
	gui.modal = modal
	return modal
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

	for i, panel := range gui.PanelNames {
		if panel == next {
			gui.active = i
		}
	}

	gui.SetNaviWithPanelName(next)
	return v
}

func (gui *Gui) IsSetView(name string) bool {
	if v, err := gui.View(name); err != nil && v == nil {
		return false
	}

	return true
}

func (gui *Gui) SetNaviWithPanelName(name string) *gocui.View {
	navi := gui.Panels[NavigatePanel].(Navigate)
	return navi.SetNavigate(name)
}

func (gui *Gui) NewFilterPanel(panel Panel, reset, closePanel func(*gocui.Gui, *gocui.View) error) error {
	maxX, maxY := gui.Size()
	x := maxX / 8
	y := maxY / 2
	w := maxX - x
	h := y + 2

	v, err := gui.SetView(FilterPanel, x, y, w, h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Title = v.Name()
		v.Wrap = true
		v.Editable = true
		v.Editor = panel
	}

	gui.SwitchPanel(v.Name())

	if err := gui.SetKeybinding(v.Name(), gocui.KeyEsc, gocui.ModNone, reset); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(v.Name(), gocui.KeyEnter, gocui.ModNone, closePanel); err != nil {
		panic(err)
	}

	return nil
}

func (gui *Gui) AddTask(taskName string, f func() error) {
	go gui.Panels[TaskListPanel].(*TaskList).StartTask(NewTask(taskName, f))
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

		line := ReadLineY(v, nexty)
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

func ReadLineY(v *gocui.View, y int) string {
	str, err := v.Line(y)

	if err != nil {
		return ""
	}

	return strings.Trim(str, " ")
}

func ReadViewBuffer(v *gocui.View) string {
	return strings.Replace(v.ViewBuffer(), "\n", "", -1)
}
