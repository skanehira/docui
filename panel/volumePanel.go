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
	name           string
	Volumes        map[string]*Volume
	Data           map[string]interface{}
	Items          Items
	ClosePanelName string
}

type Volume struct {
	Name       string `tag:"NAME" len:"min:15 max:0.3"`
	MountPoint string `tag:"MOUNTPOINT" len:"min:15 max:0.4"`
	Driver     string `tag:"DRIVER" len:"min:15 max:0.1"`
	Created    string `tag:"CREATED" len:"min:15 max:0.2"`
}

var location = time.FixedZone("Asia/Tokyo", 9*60*60)

func NewVolumeList(gui *Gui, name string, x, y, w, h int) *VolumeList {
	return &VolumeList{
		Gui:      gui,
		name:     name,
		Volumes:  make(map[string]*Volume),
		Position: Position{x, y, w, h},
		Data:     make(map[string]interface{}),
		Items:    Items{},
	}
}

func (vl *VolumeList) Name() string {
	return vl.name
}

func (vl *VolumeList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := g.SetView(VolumeListHeaderPanel, vl.x, vl.y, vl.w, vl.h); err != nil {
		if err != gocui.ErrUnknownView {
			panic(err)
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

	if err := vl.SetKeybinding(vl.name, 'j', gocui.ModNone, CursorDown); err != nil {
		panic(err)
	}
	if err := vl.SetKeybinding(vl.name, 'k', gocui.ModNone, CursorUp); err != nil {
		panic(err)
	}
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
}

func (vl *VolumeList) Refresh(g *gocui.Gui, v *gocui.View) error {
	vl.Update(func(g *gocui.Gui) error {
		v, err := vl.View(vl.name)
		if err != nil {
			panic(err)
		}

		vl.GetVolumeList(v)

		return nil
	})

	return nil
}

func (vl *VolumeList) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	return vl.Panels[vl.ClosePanelName].(*Input).ClosePanel(g, v)
}

func (vl *VolumeList) GetVolumeList(v *gocui.View) {
	v.Clear()

	var keys []string
	for _, volume := range vl.Docker.Volumes() {
		name := volume.Name
		if len(name) > 12 {
			name = name[:12]
		}

		vl.Volumes[name] = &Volume{
			Name:       volume.Name,
			MountPoint: volume.Mountpoint,
			Driver:     volume.Driver,
			Created:    volume.CreatedAt.In(location).Format("2006/01/02 15:04:05"),
		}

		keys = append(keys, name)
	}

	for _, key := range common.SortKeys(keys) {
		common.OutputFormatedLine(v, vl.Volumes[key])
	}

}

func (vl *VolumeList) CreateVolumePanel(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := vl.Size()

	x := maxX / 8
	y := maxY / 3
	w := maxX - x
	h := maxY - y

	vl.NextPanel = VolumeListPanel
	vl.ClosePanelName = CreateVolumePanel
	vl.Items = vl.NewCreateVolumeItems(x, y, w, h)

	handlers := Handlers{
		gocui.KeyEnter: vl.CreateVolume,
	}

	NewInput(vl.Gui, CreateVolumePanel, x, y, w, h, vl.Items, vl.Data, handlers)
	return nil
}

func (vl *VolumeList) CreateVolume(g *gocui.Gui, v *gocui.View) error {
	data, err := vl.GetItemsToMap(vl.Items)
	if err != nil {
		vl.ClosePanel(g, v)
		vl.ErrMessage(err.Error(), vl.NextPanel)
		return nil
	}

	options := vl.Docker.NewCreateVolumeOptions(data)

	g.Update(func(g *gocui.Gui) error {
		vl.ClosePanel(g, v)
		vl.StateMessage("volume creating...")

		g.Update(func(g *gocui.Gui) error {
			defer vl.CloseStateMessage()

			if err := vl.Docker.CreateVolumeWithOptions(options); err != nil {
				vl.ErrMessage(err.Error(), vl.NextPanel)
				return nil
			}

			vl.Panels[vl.name].Refresh(g, v)
			vl.SwitchPanel(vl.NextPanel)

			return nil
		})

		return nil
	})

	return nil
}

func (vl *VolumeList) RemoveVolume(g *gocui.Gui, v *gocui.View) error {
	name := vl.Volumes[vl.GetVolumeName(v)].Name

	if name == "" {
		return nil
	}

	vl.NextPanel = VolumeListPanel

	vl.ConfirmMessage("Are you sure you want to remove this volume? (y/n)", func(g *gocui.Gui, v *gocui.View) error {
		defer vl.Refresh(g, v)
		defer vl.CloseConfirmMessage(g, v)

		if err := vl.Docker.RemoveVolumeWithName(name); err != nil {
			vl.ErrMessage(err.Error(), vl.NextPanel)
			return nil
		}

		return nil
	})

	return nil
}

func (vl *VolumeList) PruneVolumes(g *gocui.Gui, v *gocui.View) error {

	vl.NextPanel = VolumeListPanel

	vl.ConfirmMessage("Are you sure you want to remove unused volumes? (y/n)", func(g *gocui.Gui, v *gocui.View) error {
		defer vl.Refresh(g, v)
		defer vl.CloseConfirmMessage(g, v)

		if err := vl.Docker.PruneVolumes(); err != nil {
			vl.ErrMessage(err.Error(), vl.NextPanel)
			return nil
		}

		return nil
	})

	return nil
}

func (vl *VolumeList) DetailVolume(g *gocui.Gui, v *gocui.View) error {
	name := vl.GetVolumeName(v)
	if name == "" {
		return nil
	}

	name = vl.Volumes[name].Name

	volume, err := vl.Docker.InspectVolumeWithName(name)
	if err != nil {
		panic(err)
	}

	vl.PopupDetailPanel(g, v)

	v, err = g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)

	fmt.Fprint(v, common.StructToJson(volume))

	return nil
}

func (vl *VolumeList) GetVolumeName(v *gocui.View) string {
	line := ReadLine(v, nil)
	if line == "" {
		return line
	}

	name := strings.Split(line, " ")[0]

	if len(name) > 12 {
		name = name[:12]
	}

	return name
}

func (vl *VolumeList) NewCreateVolumeItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Name",
		"Driver",
		"Labels",
		"Options",
	}

	return NewItems(names, ix, iy, iw, ih, 10)
}
