package panel

import (
	"fmt"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

type VolumeList struct {
	*Gui
	Position
	name    string
	Volumes []*Volume
	Data    map[string]interface{}
	filter  string
	form    *Form
}

type Volume struct {
	Name       string `tag:"NAME" len:"min:0.1 max:0.2"`
	MountPoint string `tag:"MOUNTPOINT" len:"min:0.1 max:0.4"`
	Driver     string `tag:"DRIVER" len:"min:0.1 max:0.2"`
	Created    string `tag:"CREATED" len:"min:0.1 max:0.2"`
}

var location = time.FixedZone("Asia/Tokyo", 9*60*60)

func NewVolumeList(gui *Gui, name string, x, y, w, h int) *VolumeList {
	return &VolumeList{
		Gui:      gui,
		name:     name,
		Position: Position{x, y, w, h},
		Data:     make(map[string]interface{}),
	}
}

func (vl *VolumeList) Name() string {
	return vl.name
}

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

func (vl *VolumeList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := g.SetView(VolumeListHeaderPanel, vl.x, vl.y, vl.w, vl.h); err != nil {
		if err != gocui.ErrUnknownView {
			vl.Logger.Error(err)
			return err
		}

		v.Wrap = true
		v.Frame = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormatedHeader(v, &Volume{})
	}

	// set scroll panel
	v, err := g.SetView(vl.name, vl.x, vl.y+1, vl.w, vl.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			vl.Logger.Error(err)
			return err
		}

		v.Frame = false
		v.Wrap = true
		v.FgColor = gocui.ColorMagenta
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)
	}

	vl.SetKeyBinding()

	//monitoring volume interval 5s
	go func() {
		for {
			vl.Update(func(g *gocui.Gui) error {
				vl.Refresh(g, v)
				return nil
			})
			time.Sleep(5 * time.Second)
		}
	}()
	return nil
}

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

func (vl *VolumeList) Refresh(g *gocui.Gui, v *gocui.View) error {
	vl.Update(func(g *gocui.Gui) error {
		v, err := vl.View(vl.name)
		if err != nil {
			vl.Logger.Error(err)
			return nil
		}

		vl.GetVolumeList(v)

		return nil
	})

	return nil
}

func (vl *VolumeList) GetVolumeList(v *gocui.View) {
	v.Clear()
	vl.Volumes = make([]*Volume, 0)

	var keys []string
	tmpMap := make(map[string]*Volume)

	for _, volume := range vl.Docker.Volumes() {
		if vl.filter != "" {
			if strings.Index(strings.ToLower(volume.Name), strings.ToLower(vl.filter)) == -1 {
				continue
			}
		}

		tmpMap[volume.Name] = &Volume{
			Name:       volume.Name,
			MountPoint: volume.Mountpoint,
			Driver:     volume.Driver,
			Created:    volume.CreatedAt.In(location).Format("2006/01/02 15:04:05"),
		}

		keys = append(keys, volume.Name)
	}

	for _, key := range common.SortKeys(keys) {
		common.OutputFormatedLine(v, tmpMap[key])
		vl.Volumes = append(vl.Volumes, tmpMap[key])
	}

}

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

func (vl *VolumeList) CreateVolume(g *gocui.Gui, v *gocui.View) error {
	data := vl.form.GetFieldTexts()

	options := vl.Docker.NewCreateVolumeOptions(data)

	vl.form.Close(g, v)

	vl.AddTask(fmt.Sprintf("Volume create %s", data["Name"]), func() error {
		vl.Logger.Info("create volume start")
		defer vl.Logger.Info("create volume finished")

		if err := vl.Docker.CreateVolumeWithOptions(options); err != nil {
			vl.Logger.Error(err)
			return nil
		}

		return vl.Refresh(g, v)
	})

	return nil
}

func (vl *VolumeList) RemoveVolume(g *gocui.Gui, v *gocui.View) error {
	selected, err := vl.selected()
	if err != nil {
		vl.ErrMessage(err.Error(), vl.name)
		vl.Logger.Error(err)
		return nil
	}

	_, err = vl.Docker.InspectVolume(selected.Name)
	if err != nil {
		vl.ErrMessage(err.Error(), vl.name)
		vl.Logger.Error(err)
		return nil
	}

	vl.ConfirmMessage("Are you sure you want to remove this volume?", vl.name, func() error {
		vl.AddTask(fmt.Sprintf("Remove container %s", selected.Name), func() error {
			vl.Logger.Info("remove volume start")
			defer vl.Logger.Info("remove volume finished")

			if err := vl.Docker.RemoveVolumeWithName(selected.Name); err != nil {
				vl.ErrMessage(err.Error(), vl.name)
				vl.Logger.Error(err)
				return err
			}

			return vl.Refresh(g, v)
		})
		return nil
	})

	return nil
}

func (vl *VolumeList) PruneVolumes(g *gocui.Gui, v *gocui.View) error {
	if len(vl.Volumes) == 0 {
		vl.ErrMessage(common.ErrNoVolume.Error(), vl.name)
		vl.Logger.Error(common.ErrNoVolume.Error(), vl.name)
		return nil
	}

	vl.ConfirmMessage("Are you sure you want to remove unused volumes?", vl.name, func() error {
		vl.AddTask("Remove unused volume", func() error {
			vl.Logger.Info("remove unused volume start")
			defer vl.Logger.Info("remove unused volume finished")

			if err := vl.Docker.PruneVolumes(); err != nil {
				vl.ErrMessage(err.Error(), vl.name)
				vl.Logger.Error(err)
				return err
			}

			return vl.Refresh(g, v)
		})
		return nil
	})
	return nil
}

func (vl *VolumeList) DetailVolume(g *gocui.Gui, v *gocui.View) error {
	vl.Logger.Info("inspect volume start")
	defer vl.Logger.Info("inspect volume finished")

	selected, err := vl.selected()
	if err != nil {
		vl.ErrMessage(err.Error(), vl.name)
		vl.Logger.Error(err)
		return nil
	}

	volume, err := vl.Docker.InspectVolume(selected.Name)
	if err != nil {
		vl.ErrMessage(err.Error(), vl.name)
		vl.Logger.Error(err)
		return nil
	}

	vl.PopupDetailPanel(g, v)

	v, err = g.View(DetailPanel)
	if err != nil {
		vl.Logger.Error(err)
		return nil
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)

	fmt.Fprint(v, common.StructToJson(volume))

	return nil
}

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
			vl.Logger.Error(err)
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
		vl.Logger.Error(err)
		return nil
	}

	return nil
}
