package panel

import (
	"fmt"
	"log"
	"os"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
)

type ImageList struct {
	*Gui
	name string
	Position
	Images         map[string]Image
	Data           map[string]interface{}
	ClosePanelName string
}

type Image struct {
	ID      string
	Name    string
	Created string
	Size    string
}

func NewImageList(gui *Gui, name string, x, y, w, h int) ImageList {
	return ImageList{
		Gui:      gui,
		name:     name,
		Position: Position{x, y, x + w, y + h},
		Images:   make(map[string]Image),
		Data:     make(map[string]interface{}),
	}
}

func (i ImageList) Name() string {
	return i.name
}

func (i ImageList) SetView(g *gocui.Gui) error {
	v, err := g.SetView(i.Name(), i.x, i.y, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = v.Name()
		v.Wrap = true
		v.SetOrigin(0, 0)
		v.SetCursor(0, 1)
	}

	i.SetKeyBinding()
	i.GetImageList(g, v)

	return nil
}

func (i ImageList) Refresh() error {
	i.Update(func(g *gocui.Gui) error {
		v, err := i.View(ImageListPanel)
		if err != nil {
			panic(err)
		}
		i.GetImageList(g, v)
		return nil
	})

	return nil
}

func (i ImageList) SetKeyBinding() {
	i.SetKeyBindingToPanel(i.name)

	if err := i.SetKeybinding(i.name, 'j', gocui.ModNone, CursorDown); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, 'k', gocui.ModNone, CursorUp); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, gocui.KeyEnter, gocui.ModNone, i.DetailImage); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, 'o', gocui.ModNone, i.DetailImage); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, 'c', gocui.ModNone, i.CreateContainerPanel); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, 'p', gocui.ModNone, i.PullImagePanel); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, 'd', gocui.ModNone, i.RemoveImage); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, 's', gocui.ModNone, i.SaveImagePanel); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, 'i', gocui.ModNone, i.ImportImagePanel); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, gocui.KeyCtrlL, gocui.ModNone, i.LoadImagePanel); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, gocui.KeyCtrlS, gocui.ModNone, i.SearchImagePanel); err != nil {
		log.Panicln(err)
	}
}

func (i ImageList) CreateContainerPanel(g *gocui.Gui, v *gocui.View) error {
	name := i.GetImageName(v)
	if name == "" {
		return nil
	}

	i.Data = map[string]interface{}{
		"Image": name,
	}

	maxX, maxY := i.Size()
	x := maxX / 8
	y := maxY / 8
	w := maxX - x
	h := maxY - y

	i.NextPanel = ImageListPanel
	i.ClosePanelName = CreateContainerPanel

	handlers := Handlers{
		gocui.KeyEnter: i.CreateContainer,
	}

	NewInput(i.Gui, CreateContainerPanel, x, y, w, h, NewCreateContainerItems(x, y, w, h), i.Data, handlers)
	return nil
}

func (i ImageList) CreateContainer(g *gocui.Gui, v *gocui.View) error {
	data, err := i.GetItemsToMap(NewCreateContainerItems(i.x, i.y, i.w, i.h))
	if err != nil {
		i.ClosePanel(g, v)
		i.ErrMessage(err.Error(), i.NextPanel)
		return nil
	}

	options, err := i.Docker.NewContainerOptions(data)

	if err != nil {
		i.ClosePanel(g, v)
		i.ErrMessage(err.Error(), i.NextPanel)
		return nil
	}

	g.Update(func(g *gocui.Gui) error {
		i.ClosePanel(g, v)
		i.StateMessage("container creating...")

		g.Update(func(g *gocui.Gui) error {
			defer i.CloseStateMessage()

			if err := i.Docker.CreateContainerWithOptions(options); err != nil {
				i.ErrMessage(err.Error(), i.NextPanel)
				return nil
			}

			i.Panels[ContainerListPanel].Refresh()
			i.SwitchPanel(i.NextPanel)

			return nil
		})

		return nil
	})

	return nil
}

func (i ImageList) PullImagePanel(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	i.NextPanel = ImageListPanel
	i.ClosePanelName = PullImagePanel

	handlers := Handlers{
		gocui.KeyEnter: i.PullImage,
	}

	NewInput(i.Gui, PullImagePanel, x, y, w, h, NewPullImageItems(x, y, w, h), i.Data, handlers)
	return nil
}

func (i ImageList) PullImage(g *gocui.Gui, v *gocui.View) error {

	item := strings.SplitN(ReadLine(v, nil), ":", 2)

	if len(item) == 0 {
		return nil
	}

	name := item[0]
	var tag string

	if len(item) == 1 {
		tag = "latest"
	} else {
		tag = item[1]
	}

	g.Update(func(g *gocui.Gui) error {
		i.ClosePanel(g, v)
		i.StateMessage("image pulling...")

		g.Update(func(g *gocui.Gui) error {
			defer i.CloseStateMessage()

			options := docker.PullImageOptions{
				Repository: name,
				Tag:        tag,
			}

			if err := i.Docker.PullImageWithOptions(options); err != nil {
				i.ErrMessage(err.Error(), i.NextPanel)
				return nil
			}

			i.Refresh()
			i.SwitchPanel(i.NextPanel)

			return nil

		})

		return nil
	})

	return nil
}

func (i ImageList) DetailImage(g *gocui.Gui, v *gocui.View) error {

	id := i.GetImageID(v)
	if id == "" {
		return nil
	}

	img, err := i.Docker.InspectImage(id)
	if err != nil {
		return err
	}

	v, err = g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	v.Clear()
	v.SetOrigin(0, 0)
	v.SetCursor(0, 0)
	fmt.Fprint(v, StructToJson(img))

	return nil
}

func (i ImageList) SaveImagePanel(g *gocui.Gui, v *gocui.View) error {

	id := i.GetImageName(v)
	if id == "" {
		return nil
	}

	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	i.NextPanel = ImageListPanel
	i.ClosePanelName = SaveImagePanel

	i.Data = map[string]interface{}{
		"ID": id,
	}

	handlers := Handlers{
		gocui.KeyEnter: i.SaveImage,
	}

	NewInput(i.Gui, SaveImagePanel, x, y, w, h, NewSaveImageItems(x, y, w, h), i.Data, handlers)
	return nil
}

func (i ImageList) SaveImage(g *gocui.Gui, v *gocui.View) error {
	path := ReadLine(v, nil)

	if path == "" {
		return nil
	}

	g.Update(func(g *gocui.Gui) error {
		i.ClosePanel(g, v)
		i.StateMessage("image saving....")

		g.Update(func(g *gocui.Gui) error {
			defer i.CloseStateMessage()

			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
			if err != nil {
				i.ErrMessage(err.Error(), i.NextPanel)
				return nil
			}
			defer file.Close()

			options := docker.ExportImageOptions{
				Name:         i.Data["ID"].(string),
				OutputStream: file,
			}

			if err := i.Docker.SaveImageWithOptions(options); err != nil {
				i.ErrMessage(err.Error(), i.NextPanel)
				return nil
			}

			i.SwitchPanel(i.NextPanel)

			return nil
		})

		return nil
	})

	return nil
}

func (i ImageList) ImportImagePanel(g *gocui.Gui, v *gocui.View) error {

	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := maxY - y

	i.NextPanel = ImageListPanel
	i.ClosePanelName = ImportImagePanel

	handlers := Handlers{
		gocui.KeyEnter: i.ImportImage,
	}

	NewInput(i.Gui, ImportImagePanel, x, y, w, h, NewImportImageItems(x, y, w, h), i.Data, handlers)
	return nil
}

func (i ImageList) ImportImage(g *gocui.Gui, v *gocui.View) error {

	data, err := i.GetItemsToMap(NewImportImageItems(i.x, i.y, i.w, i.h))
	if err != nil {
		i.ClosePanel(g, v)
		i.ErrMessage(err.Error(), i.NextPanel)
		return nil
	}

	options := docker.ImportImageOptions{
		Repository: data["Repository"],
		Source:     data["Path"],
		Tag:        data["Tag"],
	}

	g.Update(func(g *gocui.Gui) error {
		i.ClosePanel(g, v)
		i.StateMessage("image importing....")

		g.Update(func(g *gocui.Gui) error {
			defer i.CloseStateMessage()

			if err := i.Docker.ImportImageWithOptions(options); err != nil {
				i.ErrMessage(err.Error(), i.NextPanel)
				return nil
			}

			i.Refresh()
			i.SwitchPanel(i.NextPanel)

			return nil
		})

		return nil
	})

	return nil
}

func (i ImageList) LoadImagePanel(g *gocui.Gui, v *gocui.View) error {

	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	i.NextPanel = ImageListPanel
	i.ClosePanelName = LoadImagePanel

	handlers := Handlers{
		gocui.KeyEnter: i.LoadImage,
	}

	NewInput(i.Gui, LoadImagePanel, x, y, w, h, NewLoadImageItems(x, y, w, h), i.Data, handlers)
	return nil
}

func (i ImageList) LoadImage(g *gocui.Gui, v *gocui.View) error {
	path := ReadLine(v, nil)
	if path == "" {
		return nil
	}

	g.Update(func(g *gocui.Gui) error {
		i.ClosePanel(g, v)
		i.StateMessage("image loading....")

		g.Update(func(g *gocui.Gui) error {

			defer i.CloseStateMessage()
			if err := i.Docker.LoadImageWithPath(path); err != nil {
				i.ErrMessage(err.Error(), i.NextPanel)
				return nil
			}

			i.Refresh()
			i.SwitchPanel(i.NextPanel)

			return nil
		})

		return nil
	})

	return nil
}

func (i ImageList) SearchImagePanel(g *gocui.Gui, v *gocui.View) error {
	i.NextPanel = g.CurrentView().Name()

	maxX, maxY := g.Size()
	x := maxX / 8
	y := maxY / 4
	w := maxX - x
	h := y + 2

	NewSearchImage(i.Gui, SearchImagePanel, Position{x, y, w, h})
	return nil
}

func (i ImageList) GetImageList(g *gocui.Gui, v *gocui.View) {
	v.Clear()

	format := "%-15s %-40s %-25s %-15s\n"
	fmt.Fprintf(v, format, "ID", "NAME", "CREATED", "SIZE")

	for _, image := range i.Docker.Images() {
		id := image.ID[7:19]
		name := image.RepoTags[0]
		created := ParseDateToString(image.Created)
		size := ParseSizeToString(image.Size)

		i.Images[id] = Image{
			ID:      image.ID,
			Name:    name,
			Created: created,
			Size:    size,
		}
		fmt.Fprintf(v, format, id, name, created, size)
	}
}

func (i ImageList) GetImageID(v *gocui.View) string {
	line := ReadLine(v, nil)
	if line == "" || line[:2] == "ID" {
		return ""
	}

	return strings.Split(line, " ")[0]
}

func (i ImageList) GetImageName(v *gocui.View) string {
	line := ReadLine(v, nil)
	if line == "" || line[:2] == "ID" {
		return ""
	}

	image := i.Images[i.GetImageID(v)]

	return image.Name
}

func (i ImageList) RemoveImage(g *gocui.Gui, v *gocui.View) error {
	i.NextPanel = ImageListPanel
	name := i.GetImageID(v)
	if name == "" {
		return nil
	}

	i.ConfirmMessage("Do you want delete this image? (y/n)", func(g *gocui.Gui, v *gocui.View) error {
		defer i.Refresh()
		defer i.CloseConfirmMessage(g, v)

		if err := i.Docker.RemoveImageWithName(name); err != nil {
			i.ErrMessage(err.Error(), ImageListPanel)
			return nil
		}

		return nil
	})

	return nil
}

func (i ImageList) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	return i.Panels[i.ClosePanelName].(Input).ClosePanel(g, v)
}

func NewExportImageItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Path",
	}

	return NewItems(names, ix, iy, iw, ih, 6)
}

func NewSaveImageItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Path",
	}

	return NewItems(names, ix, iy, iw, ih, 6)
}

func NewImportImageItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Repository",
		"Path",
		"Tag",
	}

	return NewItems(names, ix, iy, iw, ih, 12)
}

func NewLoadImageItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Path",
	}

	return NewItems(names, ix, iy, iw, ih, 6)
}

func NewPullImageItems(ix, iy, iw, ih int) Items {
	names := []string{
		"Name",
	}

	return NewItems(names, ix, iy, iw, ih, 6)
}
