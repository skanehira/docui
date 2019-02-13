package panel

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/jroimartin/gocui"
	"github.com/skanehira/docui/common"
)

const (
	// VolumeTypeBind type of volume is bind
	VolumeTypeBind = "bind"
	// VolumeTypeVolume type of volume is volume
	VolumeTypeVolume = "volume"
)

// ImageList image list panel.
type ImageList struct {
	*Gui
	name string
	Position
	Images []*Image
	Data   map[string]interface{}
	filter string
	form   *Form
	stop   chan int
}

// Image image info.
type Image struct {
	ID      string `tag:"ID" len:"min:0.1 max:0.2"`
	Repo    string `tag:"REPOSITORY" len:"min:0.1 max:0.3"`
	Tag     string `tag:"TAG" len:"min:0.1 max:0.1"`
	Created string `tag:"CREATED" len:"min:0.1 max:0.2"`
	Size    string `tag:"SIZE" len:"min:0.1 max:0.2"`
}

// NewImageList create new image list panel.
func NewImageList(gui *Gui, name string, x, y, w, h int) *ImageList {
	i := &ImageList{
		Gui:      gui,
		name:     name,
		Position: Position{x, y, w, h},
		Data:     make(map[string]interface{}),
		stop:     make(chan int, 1),
	}

	return i
}

// Name return panel name.
func (i *ImageList) Name() string {
	return i.name
}

// Edit filtering image list.
func (i *ImageList) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
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

	i.filter = ReadViewBuffer(v)

	if v, err := i.View(i.name); err == nil {
		i.GetImageList(v)
	}
}

// SetView set up image list panel.
func (i *ImageList) SetView(g *gocui.Gui) error {
	// set header panel
	if v, err := common.SetViewWithValidPanelSize(g, ImageListHeaderPanel, i.x, i.y, i.w, i.h); err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}

		v.Wrap = true
		v.Frame = true
		v.Title = v.Name()
		v.FgColor = gocui.AttrBold | gocui.ColorWhite
		common.OutputFormattedHeader(v, &Image{})
	}

	// set scroll panel
	v, err := common.SetViewWithValidPanelSize(g, i.name, i.x, i.y+1, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			common.Logger.Error(err)
			return err
		}
		v.Frame = false
		v.Wrap = true
		v.FgColor = gocui.ColorCyan
		v.SelBgColor = gocui.ColorWhite
		v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)

		i.GetImageList(v)
	}

	i.SetKeyBinding()

	// monitoring image status.
	go i.Monitoring(i.stop, i.Gui.Gui, v)
	return nil
}

// Monitoring monitoring image list.
func (i *ImageList) Monitoring(stop chan int, g *gocui.Gui, v *gocui.View) {
	common.Logger.Info("monitoring image list start")
	ticker := time.NewTicker(5 * time.Second)

LOOP:
	for {
		select {
		case <-ticker.C:
			i.Update(func(g *gocui.Gui) error {
				return i.Refresh(g, v)
			})
		case <-stop:
			ticker.Stop()
			break LOOP
		}
	}
	common.Logger.Info("monitoring image list stop")
}

// CloseView close panel
func (i *ImageList) CloseView() {
	// stop monitoring
	i.stop <- 0
	close(i.stop)
}

// Refresh update image info
func (i *ImageList) Refresh(g *gocui.Gui, v *gocui.View) error {
	i.Update(func(g *gocui.Gui) error {
		v, err := i.View(i.name)
		if err != nil {
			common.Logger.Error(err)
			return nil
		}
		i.GetImageList(v)
		return nil
	})

	return nil
}

// SetKeyBinding set key bind to this panel.
func (i *ImageList) SetKeyBinding() {
	i.SetKeyBindingToPanel(i.name)

	if err := i.SetKeybinding(i.name, gocui.KeyEnter, gocui.ModNone, i.DetailImage); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, 'o', gocui.ModNone, i.DetailImage); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, 'c', gocui.ModNone, i.CreateContainerPanel); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, 'p', gocui.ModNone, i.PullImagePanel); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, 'd', gocui.ModNone, i.RemoveImage); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, gocui.KeyCtrlD, gocui.ModNone, i.RemoveDanglingImages); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, 's', gocui.ModNone, i.SaveImagePanel); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, 'i', gocui.ModNone, i.ImportImagePanel); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, gocui.KeyCtrlL, gocui.ModNone, i.LoadImagePanel); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, gocui.KeyCtrlF, gocui.ModNone, i.SearchImagePanel); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, gocui.KeyCtrlR, gocui.ModNone, i.Refresh); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(i.name, 'f', gocui.ModNone, i.Filter); err != nil {
		panic(err)
	}
}

// selected return selected image
func (i *ImageList) selected() (*Image, error) {
	v, _ := i.View(i.name)
	_, cy := v.Cursor()
	_, oy := v.Origin()

	index := oy + cy
	length := len(i.Images)

	if index >= length {
		return nil, common.ErrNoImage
	}

	return i.Images[index], nil
}

// CreateContainerPanel display create container form.
func (i *ImageList) CreateContainerPanel(g *gocui.Gui, v *gocui.View) error {
	// get image name
	name, err := i.GetImageName()
	if err != nil {
		i.ErrMessage(err.Error(), i.name)
		common.Logger.Error(err)
		return nil
	}

	// get position
	maxX, maxY := i.Size()
	x := maxX / 6
	y := maxY / 4
	w := x * 4

	labelw := 11
	fieldw := w - labelw

	// new form
	form := NewForm(g, CreateContainerPanel, x, y, w, 0)
	i.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		i.SwitchPanel(i.name)
		return nil
	})

	// add fields
	form.AddInput("Name", labelw, fieldw)

	form.AddInput("HostIP", labelw, fieldw)

	form.AddInput("HostPort", labelw, fieldw).
		AddValidate("no specified HostPort", func(value string) bool {
			port := form.GetFieldText("Port")
			if value == "" && port != "" {
				return false
			}
			return true
		})

	form.AddInput("Port", labelw, fieldw).
		AddValidate("no specified Port", func(value string) bool {
			hostPort := form.GetFieldText("HostPort")
			if value == "" && hostPort != "" {
				return false
			}
			return true
		})

	form.AddSelectOption("VolumeType", labelw, fieldw).
		AddOptions([]string{VolumeTypeBind, VolumeTypeVolume}...)

	form.AddInput("HostVolume", labelw, fieldw).
		AddValidate("no specified HostVolume", func(value string) bool {
			volume := form.GetFieldText("Volume")
			if value == "" && volume != "" {
				return false
			}
			return true
		})
	form.AddInput("Volume", labelw, fieldw).
		AddValidate("no specified Volume", func(value string) bool {
			hostVolume := form.GetFieldText("HostVolume")
			if hostVolume != "" && value == "" {
				return false
			}
			return true
		})

	form.AddInput("Image", labelw, fieldw).
		SetText(name).
		AddValidate("no specified Image", func(value string) bool {
			return value != ""
		})

	form.AddInput("User", labelw, fieldw)
	form.AddCheckBox("Attach", labelw)
	form.AddInput("Env", labelw, fieldw)
	form.AddInput("Cmd", labelw, fieldw)
	form.AddButton("Create", i.CreateContainer)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()
	return nil
}

// CreateContainer create the container.
func (i *ImageList) CreateContainer(g *gocui.Gui, v *gocui.View) error {
	if !i.form.Validate() {
		return nil
	}

	data := i.form.GetFieldTexts()
	data["VolumeType"] = i.form.GetSelectedOpt("VolumeType")

	options, err := i.Docker.NewContainerOptions(data, i.form.GetCheckBoxState("Attach"))

	if err != nil {
		i.form.Close(g, v)
		i.ErrMessage(err.Error(), i.name)
		common.Logger.Error(err)
		return nil
	}

	i.form.Close(g, v)

	i.AddTask(fmt.Sprintf("Create container %s", data["Name"]), func() error {
		common.Logger.Info("create image start")
		defer common.Logger.Info("create image end")

		if err := i.Docker.CreateContainer(options); err != nil {
			common.Logger.Error(err)
			return err
		}

		return i.Panels[ContainerListPanel].Refresh(g, v)
	})

	return nil
}

// PullImagePanel display pull image form.
func (i *ImageList) PullImagePanel(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()
	x := maxX / 8
	y := maxY / 3
	w := x * 6

	labelw := 6
	fieldw := w - labelw

	// new form
	form := NewForm(g, PullImagePanel, x, y, w, 0)
	i.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		i.SwitchPanel(i.name)
		return nil
	})

	// add fields
	form.AddInput("Image", labelw, fieldw).
		AddValidate(Require.Message+"Image", Require.Validate)
	form.AddButton("Pull", i.PullImage)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()
	return nil
}

// PullImage pull the specified image.
func (i *ImageList) PullImage(g *gocui.Gui, v *gocui.View) error {
	if !i.form.Validate() {
		return nil
	}

	i.form.Close(g, v)

	image := i.form.GetFieldTexts()["Image"]
	i.AddTask(fmt.Sprintf("Pull image %s", image), func() error {
		common.Logger.Info("pull image start")
		defer common.Logger.Info("pull image end")

		if err := i.Docker.PullImage(image); err != nil {
			common.Logger.Error(err)
			return err
		}

		return i.Refresh(g, v)
	})

	return nil
}

// DetailImage display the image detail info
func (i *ImageList) DetailImage(g *gocui.Gui, v *gocui.View) error {
	common.Logger.Info("inspect image start")
	defer common.Logger.Info("inspect image end")

	image, err := i.selected()
	if err != nil {
		i.ErrMessage(err.Error(), i.name)
		common.Logger.Error(err)
		return nil
	}

	img, err := i.Docker.InspectImage(image.ID)
	if err != nil {
		i.ErrMessage(err.Error(), i.name)
		common.Logger.Error(err)
		return nil
	}

	i.PopupDetailPanel(g, v)

	v, err = g.View(DetailPanel)
	if err != nil {
		common.Logger.Error(err)
		return nil
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)
	fmt.Fprint(v, common.StructToJSON(img))

	return nil
}

// SaveImagePanel display save image form.
func (i *ImageList) SaveImagePanel(g *gocui.Gui, v *gocui.View) error {
	name, err := i.GetImageName()
	if err != nil {
		i.ErrMessage(err.Error(), i.name)
		common.Logger.Error(err)
		return nil
	}

	maxX, maxY := i.Size()
	x := maxX / 8
	y := maxY / 3
	w := x * 6

	labelw := 6
	fieldw := w - labelw

	// new form
	form := NewForm(g, SaveImagePanel, x, y, w, 0)
	i.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		i.SwitchPanel(i.name)
		return nil
	})

	// add fields
	form.AddInput("Path", labelw, fieldw).
		AddValidate(Require.Message+"Path", Require.Validate)
	form.AddInput("Image", labelw, fieldw).
		AddValidate(Require.Message+"Image", Require.Validate).
		SetText(name)
	form.AddButton("Save", i.SaveImage)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()

	return nil
}

// SaveImage save then specified image.
func (i *ImageList) SaveImage(g *gocui.Gui, v *gocui.View) error {

	if !i.form.Validate() {
		return nil
	}
	data := i.form.GetFieldTexts()

	i.form.Close(g, v)

	i.AddTask(fmt.Sprintf("Save image:%s to %s", data["Image"], data["Path"]), func() error {
		common.Logger.Info("save image start")
		defer common.Logger.Info("save image end")

		return i.Docker.SaveImage([]string{data["Image"]}, data["Path"])
	})

	return nil
}

// ImportImagePanel display import form.
func (i *ImageList) ImportImagePanel(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()
	x := maxX / 8
	y := maxY / 3
	w := x * 6

	labelw := 11
	fieldw := w - labelw

	// new form
	form := NewForm(g, ImportImagePanel, x, y, w, 0)
	i.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		i.SwitchPanel(i.name)
		return nil
	})

	// add fields
	form.AddInput("Repository", labelw, fieldw).
		AddValidate(Require.Message+"Repository", Require.Validate)
	form.AddInput("Path", labelw, fieldw).
		AddValidate(Require.Message+"Path", Require.Validate)
	form.AddInput("Tag", labelw, fieldw)
	form.AddButton("Import", i.ImportImage)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()

	return nil
}

// ImportImage import the specified file path.
func (i *ImageList) ImportImage(g *gocui.Gui, v *gocui.View) error {

	if !i.form.Validate() {
		return nil
	}

	data := i.form.GetFieldTexts()

	i.form.Close(g, v)

	i.AddTask(fmt.Sprintf("Import image from %s", data["Path"]), func() error {
		common.Logger.Info("import image start")
		defer common.Logger.Info("import image end")

		if err := i.Docker.ImportImage(data["Repository"], data["Tag"], data["Path"]); err != nil {
			common.Logger.Error(err)
			return err
		}

		return i.Refresh(g, v)
	})

	return nil
}

// LoadImagePanel disply load image form.
func (i *ImageList) LoadImagePanel(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()
	x := maxX / 8
	y := maxY / 3
	w := x * 6

	labelw := 6
	fieldw := w - labelw

	// new form
	form := NewForm(g, LoadImagePanel, x, y, w, 0)
	i.form = form

	// add func do after close
	form.AddCloseFunc(func() error {
		i.SwitchPanel(i.name)
		return nil
	})

	// add fields
	form.AddInput("Path", labelw, fieldw).
		AddValidate(Require.Message+"Path", Require.Validate)
	form.AddButton("Load", i.LoadImage)
	form.AddButton("Cancel", form.Close)

	// draw form
	form.Draw()
	return nil
}

// LoadImage load the specified file path.
func (i *ImageList) LoadImage(g *gocui.Gui, v *gocui.View) error {

	if !i.form.Validate() {
		return nil
	}

	path := i.form.GetFieldTexts()["Path"]

	i.form.Close(g, v)

	i.AddTask(fmt.Sprintf("Load image from %s", path), func() error {
		common.Logger.Info("load image start")
		defer common.Logger.Info("load image end")

		if err := i.Docker.LoadImage(path); err != nil {
			common.Logger.Error(err)
			return err
		}

		return i.Refresh(g, v)
	})

	return nil
}

// SearchImagePanel display the search form.
func (i *ImageList) SearchImagePanel(g *gocui.Gui, v *gocui.View) error {
	i.name = g.CurrentView().Name()

	maxX, maxY := g.Size()
	x := maxX / 8
	y := maxY / 4
	w := maxX - x
	h := y + 2

	NewSearchImage(i.Gui, SearchImagePanel, Position{x, y, w, h})
	return nil
}

// GetImageList return images info
func (i *ImageList) GetImageList(v *gocui.View) {
	v.Clear()
	i.Images = make([]*Image, 0)

	images, err := i.Docker.Images(types.ImageListOptions{
		All: true,
	})

	if err != nil {
		common.Logger.Error(err)
		return
	}

	for _, image := range images {
		for _, repoTag := range image.RepoTags {
			repo, tag := common.ParseRepoTag(repoTag)

			if i.filter != "" {
				name := fmt.Sprintf("%s:%s", repo, tag)
				if strings.Index(strings.ToLower(name), strings.ToLower(i.filter)) == -1 {
					continue
				}
			}

			id := image.ID[7:19]
			created := common.ParseDateToString(image.Created)
			size := common.ParseSizeToString(image.Size)

			image := &Image{
				ID:      id,
				Repo:    repo,
				Tag:     tag,
				Created: created,
				Size:    size,
			}

			i.Images = append(i.Images, image)

			common.OutputFormattedLine(v, image)
		}
	}
}

// GetImageName return the specified image name
func (i *ImageList) GetImageName() (string, error) {
	image, err := i.selected()
	if err != nil {
		common.Logger.Error(err)
		return "", err
	}

	var name string
	if image.Repo == "<none>" || image.Tag == "<none>" {
		name = image.ID
	} else {
		name = fmt.Sprintf("%s:%s", image.Repo, image.Tag)
	}

	return name, nil
}

// RemoveImage remove the specified image
func (i *ImageList) RemoveImage(g *gocui.Gui, v *gocui.View) error {
	name, err := i.GetImageName()
	if err != nil {
		i.ErrMessage(err.Error(), i.name)
		common.Logger.Error(err)
		return nil
	}

	i.ConfirmMessage("Are you sure you want to remove this image?", i.name, func() error {
		i.AddTask(fmt.Sprintf("Remove image %s", name), func() error {
			common.Logger.Info("remove image start")
			defer common.Logger.Info("remove image end")

			if err := i.Docker.RemoveImage(name); err != nil {
				i.ErrMessage(err.Error(), i.name)
				common.Logger.Error(err)
				return err
			}

			return i.Refresh(g, v)
		})

		return nil
	})

	return nil
}

// RemoveDanglingImages remove then dangling images.
func (i *ImageList) RemoveDanglingImages(g *gocui.Gui, v *gocui.View) error {
	if len(i.Images) == 0 {
		i.ErrMessage(common.ErrNoImage.Error(), i.name)
		return nil
	}

	i.ConfirmMessage("Are you sure you want to remove unused images?", i.name, func() error {
		i.AddTask("Remove unused image", func() error {
			common.Logger.Info("remove unused image start")
			defer common.Logger.Info("remove unused image end")

			if err := i.Docker.RemoveDanglingImages(); err != nil {
				i.ErrMessage(err.Error(), i.name)
				common.Logger.Error(err)
				return err
			}
			return i.Refresh(g, v)
		})

		return nil
	})

	return nil
}

// Filter display filtering form.
func (i *ImageList) Filter(g *gocui.Gui, lv *gocui.View) error {
	isReset := false
	closePanel := func(g *gocui.Gui, v *gocui.View) error {
		if isReset {
			i.filter = ""
		} else {
			lv.SetCursor(0, 0)
			i.filter = ReadViewBuffer(v)
		}
		if v, err := i.View(i.name); err == nil {
			i.GetImageList(v)
		}

		if err := g.DeleteView(v.Name()); err != nil {
			common.Logger.Error(err)
			return nil
		}

		g.DeleteKeybindings(v.Name())
		i.SwitchPanel(i.name)
		return nil
	}

	reset := func(g *gocui.Gui, v *gocui.View) error {
		isReset = true
		return closePanel(g, v)
	}

	if err := i.NewFilterPanel(i, reset, closePanel); err != nil {
		common.Logger.Error(err)
		return nil
	}

	return nil
}
