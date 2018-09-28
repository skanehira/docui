package panel

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jroimartin/gocui"
)

type VolumeList struct {
	*Gui
	Position
	name           string
	Volumes        map[string]Volume
	Data           map[string]interface{}
	ClosePanelName string
}

type Volume struct {
	Name       string
	MountPoint string
	Driver     string
	Created    string
}

var location = time.FixedZone("Asia/Tokyo", 9*60*60)

func NewVolumeList(gui *Gui, name string, x, y, w, h int) VolumeList {
	return VolumeList{
		Gui:      gui,
		name:     name,
		Volumes:  make(map[string]Volume),
		Position: Position{x, y, w + x, y + h},
		Data:     make(map[string]interface{}),
	}
}

func (vl VolumeList) Name() string {
	return vl.name
}

func (vl VolumeList) SetView(g *gocui.Gui) error {
	v, err := g.SetView(vl.name, vl.x, vl.y, vl.w, vl.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = v.Name()
		v.Wrap = true
		v.SetOrigin(0, 0)
		v.SetCursor(0, 1)
	}

	vl.SetKeyBinding()
	vl.GetVolumeList(v)

	return nil
}

func (vl VolumeList) SetKeyBinding() {
	vl.SetKeyBindingToPanel(vl.name)

	if err := vl.SetKeybinding(vl.name, 'j', gocui.ModNone, CursorDown); err != nil {
		log.Panicln(err)
	}
	if err := vl.SetKeybinding(vl.name, 'k', gocui.ModNone, CursorUp); err != nil {
		log.Panicln(err)
	}
}

func (vl VolumeList) Refresh() error {
	vl.Update(func(g *gocui.Gui) error {
		v, err := vl.View(ContainerListPanel)
		if err != nil {
			panic(err)
		}

		vl.GetVolumeList(v)

		return nil
	})

	return nil
}

func (vl VolumeList) GetVolumeList(v *gocui.View) {
	v.Clear()

	c1, c2, c3, c4 := 20, 30, 15, 20

	format := "%-" + strconv.Itoa(c1) + "s %-" + strconv.Itoa(c2) + "s %-" + strconv.Itoa(c3) + "s %-" + strconv.Itoa(c4) + "s\n"
	fmt.Fprintf(v, format, "NAME", "MOUNTPOINT", "DRIVER", "CREATED")

	for _, volume := range vl.Docker.Volumes() {
		name := volume.Name
		mountPoint := volume.Mountpoint
		driver := volume.Driver

		created := volume.CreatedAt.In(location).Format("2006/01/02 15:04:05")

		vl.Volumes[name] = Volume{
			Name:       volume.Name,
			MountPoint: volume.Mountpoint,
			Driver:     volume.Driver,
			Created:    created,
		}

		if len(name) > c1 {
			name = name[:c1-3] + "..."
		}
		if len(mountPoint) > c2 {
			mountPoint = mountPoint[:c2-3] + "..."
		}
		if len(driver) > c3 {
			driver = driver[:c3-3] + "..."
		}
		if len(created) > c4 {
			created = created[:c4-3] + "..."
		}

		fmt.Fprintf(v, format, name, mountPoint, driver, created)
	}
}
