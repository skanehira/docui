package panel

import (
	"fmt"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

// VolumeList volume list panel.
type VolumeList struct {
	*Gui
	Position
	name    string
	Volumes []*Volume
	Data    map[string]interface{}
	filter  string
	form    *Form
	stop    chan int
}

// Volume volume info
type Volume struct {
	Name       string `tag:"NAME" len:"min:0.1 max:0.2"`
	MountPoint string `tag:"MOUNTPOINT" len:"min:0.1 max:0.4"`
	Driver     string `tag:"DRIVER" len:"min:0.1 max:0.2"`
	Created    string `tag:"CREATED" len:"min:0.1 max:0.2"`
}

// replace date format
var replacer = strings.NewReplacer("T", " ", "Z", "")

// NewVolumeList create new volume list panel.
func NewVolumeList(gui *Gui, name string, x, y, w, h int) *VolumeList {
	return &VolumeList{
		Gui:      gui,
		name:     name,
		Position: Position{x, y, w, h},
		Data:     make(map[string]interface{}),
		stop:     make(chan int, 1),
	}
}

// Name return panel name.
func (vl *VolumeList) Name() string {
	return vl.name
}

// Edit filtering volume list.
func (vl *VolumeList) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
		return
	case key == gocui.KeyArrowRight:
		v.MoveCursor(+1, 0, false)
		return
	}

	vl.filter = ReadViewBuffer(v)

	if v, err := vl.View(vl.name); err == nil {
		vl.GetVolumeList(v)
	}
}

// SetView set up volume list panel.
func (vl *VolumeList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := common.SetViewWithValidPanelSize(g, VolumeListHeaderPanel, vl.x, vl.y, vl.w, vl.h); err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}

		v.Wrap = true
		v.Frame = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormattedHeader(v, &Volume{})
	}

	// set scroll panel
	v, err := common.SetViewWithValidPanelSize(g, vl.name, vl.x, vl.y+1, vl.w, vl.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}

		v.Frame = false
		v.Wrap = true
		v.FgColor = gocui.ColorMagenta
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)

		vl.GetVolumeList(v)
	}

	vl.SetKeyBinding()

	// monitoring volume status.
	go vl.Monitoring(vl.stop, vl.Gui.Gui, v)
	return nil
}

// Monitoring monitoring image list.
func (vl *VolumeList) Monitoring(stop chan int, g *gocui.Gui, v *gocui.View) {
	common.Logger.Info("monitoring volume list start")
	ticker := time.NewTicker(5 * time.Second)

LOOP:
	for {
		select {
		case <-ticker.C:
			vl.Update(func(g *gocui.Gui) error {
				return vl.Refresh(g, v)
			})
		case <-stop:
			ticker.Stop()
			break LOOP
		}
	}
	common.Logger.Info("monitoring volume list stop")
}

// CloseView close panel
func (vl *VolumeList) CloseView() {
	// stop monitoring
	vl.stop <- 0
	close(vl.stop)
}

// SetKeyBinding set key bind to this panel.
func (vl *VolumeList) SetKeyBinding() {
	vl.SetKeyBindingToPanel(vl.name)

	if err := vl.SetKeybinding(vl.name, 'c', gocui.ModNone, vl.CreateVolumePanel); err != nil {
		panic(err)
	}
	if err := vl.SetKeybinding(vl.name, 'd', gocui.ModNone, vl.RemoveVolume); err != nil {
		panic(err)
	}
	if err := vl.SetKeybinding(vl.name, 'p', gocui.ModNone, vl.PruneVolumes); err != nil {
		panic(err)
	}
	if err := vl.SetKeybinding(vl.name, 'o', gocui.ModNone, vl.DetailVolume); err != nil {
		panic(err)
	}
	if err := vl.SetKeybinding(vl.name, gocui.KeyEnter, gocui.ModNone, vl.DetailVolume); err != nil {
		panic(err)
	}
	if err := vl.SetKeybinding(vl.name, gocui.KeyCtrlR, gocui.ModNone, vl.Refresh); err != nil {
		panic(err)
	}
	if err := vl.SetKeybinding(vl.name, 'f', gocui.ModNone, vl.Filter); err != nil {
		panic(err)
	}
}

// selected return selected volume.
func (vl *VolumeList) selected() (*Volume, error) {
	v, _ := vl.View(vl.name)
	_, cy := v.Cursor()
	_, oy := v.Origin()

	index := oy + cy
	length := len(vl.Volumes)

	if index >= length {
		return nil, common.ErrNoVolume
	}
	return vl.Volumes[cy+oy], nil
}

// Refresh update volume info.
func (vl *VolumeList) Refresh(g *gocui.Gui, v *gocui.View) error {
	vl.Update(func(g *gocui.Gui) error {
		v, err := vl.View(vl.name)
		if err != nil {
			common.Logger.Error(err)
			return nil
		}

		vl.GetVolumeList(v)

		return nil
	})

	return nil
}

// GetVolumeList return volumes
func (vl *VolumeList) GetVolumeList(v *gocui.View) {
	v.Clear()
	vl.Volumes = make([]*Volume, 0)

	volumes, err := vl.Docker.Volumes()

	if err != nil {
		common.Logger.Error(err)
		return
	}

	keys := make([]string, 0, len(volumes))
	tmpMap := make(map[string]*Volume)

	for _, volume := range volumes {
		if vl.filter != "" {
			if strings.Index(strings.ToLower(volume.Name), strings.ToLower(vl.filter)) == -1 {
				continue
			}
		}

		tmpMap[volume.Name] = &Volume{
			Name:       volume.Name,
			MountPoint: volume.Mountpoint,
			Driver:     volume.Driver,
			Created:    replacer.Replace(volume.CreatedAt),
		}

		keys = append(keys, volume.Name)
	}

	for _, key := range common.SortKeys(keys) {
		common.OutputFormattedLine(v, tmpMap[key])
		vl.Volumes = append(vl.Volumes, tmpMap[key])
	}
}

// CreateVolumePanel display create volume form.
func (vl *VolumeList) CreateVolumePanel(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := vl.Size()

	x := maxX / 8
	y := maxY / 3
	w := x * 6

	labelw := 8
	fieldw := w - labelw

	// new form
	form := NewForm(g, CreateVolumePanel, x, y, w, 0)
	vl.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		vl.SwitchPanel(vl.name)
		return nil
	})

	// add fields
	form.AddInput("Name", labelw, fieldw)
	form.AddInput("Driver", labelw, fieldw)
	form.AddInput("Labels", labelw, fieldw)
	form.AddInput("Options", labelw, fieldw)
	form.AddButton("Create", vl.CreateVolume)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()
	return nil
}

// CreateVolume create volume
func (vl *VolumeList) CreateVolume(g *gocui.Gui, v *gocui.View) error {
	data := vl.form.GetFieldTexts()

	vl.form.Close(g, v)

	vl.AddTask(fmt.Sprintf("Volume create %s", data["Name"]), func() error {
		common.Logger.Info("create volume start")
		defer common.Logger.Info("create volume end")

		if err := vl.Docker.CreateVolume(vl.Docker.NewCreateVolumeOptions(data)); err != nil {
			common.Logger.Error(err)
			return nil
		}

		return vl.Refresh(g, v)
	})

	return nil
}

// RemoveVolume remove the specified volume
func (vl *VolumeList) RemoveVolume(g *gocui.Gui, v *gocui.View) error {
	selected, err := vl.selected()
	if err != nil {
		vl.ErrMessage(err.Error(), vl.name)
		common.Logger.Error(err)
		return nil
	}

	_, err = vl.Docker.InspectVolume(selected.Name)
	if err != nil {
		vl.ErrMessage(err.Error(), vl.name)
		common.Logger.Error(err)
		return nil
	}

	vl.ConfirmMessage("Are you sure you want to remove this volume?", vl.name, func() error {
		vl.AddTask(fmt.Sprintf("Remove container %s", selected.Name), func() error {
			common.Logger.Info("remove volume start")
			defer common.Logger.Info("remove volume end")

			if err := vl.Docker.RemoveVolume(selected.Name); err != nil {
				vl.ErrMessage(err.Error(), vl.name)
				common.Logger.Error(err)
				return err
			}

			return vl.Refresh(g, v)
		})
		return nil
	})

	return nil
}

// PruneVolumes remove unused volumes.
func (vl *VolumeList) PruneVolumes(g *gocui.Gui, v *gocui.View) error {
	if len(vl.Volumes) == 0 {
		vl.ErrMessage(common.ErrNoVolume.Error(), vl.name)
		common.Logger.Error(common.ErrNoVolume)
		return nil
	}

	vl.ConfirmMessage("Are you sure you want to remove unused volumes?", vl.name, func() error {
		vl.AddTask("Remove unused volume", func() error {
			common.Logger.Info("remove unused volume start")
			defer common.Logger.Info("remove unused volume end")

			if err := vl.Docker.PruneVolumes(); err != nil {
				vl.ErrMessage(err.Error(), vl.name)
				common.Logger.Error(err)
				return err
			}

			return vl.Refresh(g, v)
		})
		return nil
	})
	return nil
}

// DetailVolume display detail the specified volume.
func (vl *VolumeList) DetailVolume(g *gocui.Gui, v *gocui.View) error {
	common.Logger.Info("inspect volume start")
	defer common.Logger.Info("inspect volume end")

	selected, err := vl.selected()
	if err != nil {
		vl.ErrMessage(err.Error(), vl.name)
		common.Logger.Error(err)
		return nil
	}

	volume, err := vl.Docker.InspectVolume(selected.Name)
	if err != nil {
		vl.ErrMessage(err.Error(), vl.name)
		common.Logger.Error(err)
		return nil
	}

	vl.PopupDetailPanel(g, v)

	v, err = g.View(DetailPanel)
	if err != nil {
		common.Logger.Error(err)
		return nil
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)

	fmt.Fprint(v, common.StructToJSON(volume))

	return nil
}

// Filter filtering volume
func (vl *VolumeList) Filter(g *gocui.Gui, lv *gocui.View) error {
	isReset := false
	closePanel := func(g *gocui.Gui, v *gocui.View) error {
		if isReset {
			vl.filter = ""
		} else {
			lv.SetCursor(0, 0)
			vl.filter = ReadViewBuffer(v)
		}
		if v, err := vl.View(vl.name); err == nil {
			vl.GetVolumeList(v)
		}

		if err := g.DeleteView(v.Name()); err != nil {
			common.Logger.Error(err)
			return nil
		}

		g.DeleteKeybindings(v.Name())
		vl.SwitchPanel(vl.name)
		return nil
	}

	reset := func(g *gocui.Gui, v *gocui.View) error {
		isReset = true
		return closePanel(g, v)
	}

	if err := vl.NewFilterPanel(vl, reset, closePanel); err != nil {
		common.Logger.Error(err)
		return nil
	}

	return nil
}
