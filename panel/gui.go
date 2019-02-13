package panel

import (
	"errors"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/docker"
	component "github.com/skanehira/gocui-component"
)

const (
	// ImageListPanel image panel scroll name.
	ImageListPanel = "image list scroll"
	// ImageListHeaderPanel image panel header name.
	ImageListHeaderPanel = "image list"
	// PullImagePanel pull image panel name.
	PullImagePanel = "pull image"
	// ContainerListPanel container panel scroll name.
	ContainerListPanel = "container list scroll"
	// ContainerListHeaderPanel container panel header name.
	ContainerListHeaderPanel = "container list"
	// DetailPanel detail panel name.
	DetailPanel = "detail"
	// CreateContainerPanel create container panel name.
	CreateContainerPanel = "create container"
	// SaveImagePanel save image panel name.
	SaveImagePanel = "save image"
	// ImportImagePanel import image panel name.
	ImportImagePanel = "import image"
	// LoadImagePanel load image panel name.
	LoadImagePanel = "load image"
	// ExportContainerPanel export container panel name.
	ExportContainerPanel = "export container"
	// CommitContainerPanel commit container panel name.
	CommitContainerPanel = "commit container"
	// RenameContainerPanel rename container panel name.
	RenameContainerPanel = "rename container"
	// SearchImagePanel search image panel name.
	SearchImagePanel = "search images"
	// SearchImageResultPanel search image result panel name.
	SearchImageResultPanel = "images scroll"
	// SearchImageResultHeaderPanel search image result panel header name.
	SearchImageResultHeaderPanel = "images"
	// VolumeListPanel volume list panel name.
	VolumeListPanel = "volume list scroll"
	// VolumeListHeaderPanel volume list panel header name.
	VolumeListHeaderPanel = "volume list"
	// CreateVolumePanel create volume panel name.
	CreateVolumePanel = "create volume"
	// NavigatePanel navigate panel name.
	NavigatePanel = "navigate"
	// InfoPanel info panel name.
	InfoPanel = "info"
	// DockerInfoPanel docker info panel
	DockerInfoPanel = "docker info"
	// FilterPanel filter panel name.
	FilterPanel = "filter"
	// NetworkListPanel network list panel name
	NetworkListPanel = "network list scroll"
	// NetworkListHeaderPanel network list panel header name.
	NetworkListHeaderPanel = "network list"
	// TaskListHeaderPanel task list panel header name
	TaskListHeaderPanel = "task list"
	// TaskListPanel task list panel name.
	TaskListPanel = "task list scroll"
	// ExecContainerCmd exec command panel name.
	ExecContainerCmd = "exec container cmd"
)

// ErrExecFlag use to attach container
// TODO improvement this logic
var ErrExecFlag = errors.New("exec")

// Gui have panels and docker client, logger, etc...
// The fields here can be used in the panel.
type Gui struct {
	*gocui.Gui
	Docker     *docker.Docker
	Panels     map[string]Panel
	PanelNames []string
	NextPanel  string
	active     int
	modal      *component.Modal
}

// Panel is a interface.
// It is necessary to implement it when adding a new panel.
type Panel interface {
	SetView(*gocui.Gui) error
	CloseView()
	Name() string
	Refresh(*gocui.Gui, *gocui.View) error
	Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier)
}

// Position panel must ve have position.
type Position struct {
	x, y int
	w, h int
}

// New new gui
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
	}

	gui.init()

	return gui
}

// Close close gui and logger.
func (gui *Gui) Close() {
	gui.Gui.Close()
	for _, panel := range gui.Panels {
		panel.CloseView()
	}
}

// AddPanelNames add panel name to switch any panels.
func (gui *Gui) AddPanelNames(panel Panel) {
	name := panel.Name()
	gui.PanelNames = append(gui.PanelNames, name)
}

// SetKeyBindingToPanel set key bind to any panels.
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
	if err := gui.SetKeybinding(panel, 'k', gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
	if err := gui.SetKeybinding(panel, gocui.KeyArrowUp, gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
}

// SetGlobalKeyBinding set key bind to all panels.
func (gui *Gui) SetGlobalKeyBinding() {
	if err := gui.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, gui.quit); err != nil {
		panic(err)
	}
}

// nextPanel move next panel.
func (gui *Gui) nextPanel(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (gui.active + 1) % len(gui.PanelNames)
	name := gui.PanelNames[nextIndex]

	gui.SwitchPanel(name)
	gui.active = nextIndex
	return nil
}

// prePanel move previous panel.
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

// when this is called, docui exits
func (gui *Gui) quit(g *gocui.Gui, v *gocui.View) error {
	for _, task := range gui.Panels[TaskListPanel].(*TaskList).ViewTask {
		if task.Status == Executing.String() {
			gui.ErrMessage("task is running", gui.PanelNames[gui.active])
			return nil
		}
	}
	return gocui.ErrQuit
}

// init add panel struct to gui and draw panel view
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

// StorePanels add panel name to switch panels.
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

// PopupDetailPanel display detail panel.
func (gui *Gui) PopupDetailPanel(g *gocui.Gui, v *gocui.View) error {
	gui.NextPanel = g.CurrentView().Name()

	maxX, maxY := g.Size()
	panel := NewDetail(gui, DetailPanel, 0, 0, maxX-1, maxY-3)

	panel.SetView(g)

	return nil
}

// ErrMessage display err message dialog.
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

// ConfirmMessage display confirm message dialog.
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

// StateMessage display any message in dialog.
func (gui *Gui) StateMessage(message string) {
	gui.NewModal(message)
}

// CloseStateMessage  close state message dialog.
func (gui *Gui) CloseStateMessage() {
	gui.modal.Close()
}

// NewModal create modal
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

// RefreshAllPanel refresh all panel status
func (gui *Gui) RefreshAllPanel() {
	for _, panel := range gui.Panels {
		v, _ := gui.View(panel.Name())
		panel.Refresh(gui.Gui, v)
	}

	gui.SwitchPanel(gui.NextPanel)
}

// SwitchPanel switch specific panel
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

// IsSetView Check if the panel's view has been added to gocui
func (gui *Gui) IsSetView(name string) bool {
	if v, err := gui.View(name); err != nil && v == nil {
		return false
	}

	return true
}

// SetNaviWithPanelName set navi panel message
func (gui *Gui) SetNaviWithPanelName(name string) *gocui.View {
	navi := gui.Panels[NavigatePanel].(Navigate)
	return navi.SetNavigate(name)
}

// NewFilterPanel display filter input field.
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

// AddTask add task
func (gui *Gui) AddTask(taskName string, f func() error) {
	go gui.Panels[TaskListPanel].(*TaskList).StartTask(NewTask(taskName, f))
}

// SetCurrentPanel switch panel
func SetCurrentPanel(g *gocui.Gui, name string) (*gocui.View, error) {
	v, err := g.SetCurrentView(name)

	if err != nil {
		return nil, err
	}

	v.Highlight = true

	return g.SetViewOnTop(name)
}

// CursorDown move the cursor down.
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

// CursorUp move the cursor up.
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

// PageDown move the cursor down of screen half
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

// PageUp move the cursor up of screen half
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

// ReadLineY read the line of specific y position.
func ReadLineY(v *gocui.View, y int) string {
	str, err := v.Line(y)

	if err != nil {
		return ""
	}

	return strings.Trim(str, " ")
}

// ReadViewBuffer read line.
func ReadViewBuffer(v *gocui.View) string {
	return strings.Replace(v.ViewBuffer(), "\n", "", -1)
}
