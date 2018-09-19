package panel

import (
	"fmt"
	"log"
	"strings"

	"github.com/jroimartin/gocui"
)

type ImageList struct {
	*Gui
	name string
	Position
	Images map[string]Image
}

type Image struct {
	ID      string
	Name    string
	Created string
	Size    string
}

func NewImageList(gui *Gui, name string, x, y, w, h int) ImageList {
	return ImageList{gui, name, Position{x, y, x + w, y + h}, make(map[string]Image)}
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
		if err := i.SetView(g); err != nil {
			return err
		}
		return nil
	})

	return nil
}

func (i ImageList) SetKeyBinding() {
	// keybinding
	i.DeleteKeybindings(i.name)
	i.SetKeybinds(i.name)

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
	if err := i.SetKeybinding(i.name, 's', gocui.ModNone, i.SaveImage); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, 'i', gocui.ModNone, i.ImportImage); err != nil {
		log.Panicln(err)
	}
	if err := i.SetKeybinding(i.name, gocui.KeyCtrlL, gocui.ModNone, i.LoadImage); err != nil {
		log.Panicln(err)
	}

}

func (i ImageList) CreateContainerPanel(g *gocui.Gui, v *gocui.View) error {
	i.NextPanel = ImageListPanel
	id := i.GetImageID(v)
	if id == "" {
		return nil
	}

	data := map[string]interface{}{
		"Image": id,
	}

	maxX, maxY := i.Size()
	x := maxX / 8
	y := maxY / 8
	w := maxX - x
	h := maxY - y

	NewInput(i.Gui, CreateContainerPanel, x, y, w, h, NewCreateContainerItems(x, y, w, h), data)
	return nil
}

func (i ImageList) PullImagePanel(g *gocui.Gui, v *gocui.View) error {
	i.NextPanel = ImageListPanel
	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	NewInput(i.Gui, PullImagePanel, x, y, w, h, NewPullImageItems(x, y, w, h), make(map[string]interface{}))
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

func (i ImageList) SaveImage(g *gocui.Gui, v *gocui.View) error {
	i.NextPanel = ImageListPanel

	id := i.GetImageName(v)
	if id == "" {
		return nil
	}

	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	data := map[string]interface{}{
		"ID": id,
	}

	NewInput(i.Gui, SaveImagePanel, x, y, w, h, NewExportImageItems(x, y, w, h), data)
	return nil
}

func (i ImageList) ImportImage(g *gocui.Gui, v *gocui.View) error {
	i.NextPanel = ImageListPanel

	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := maxY - y

	NewInput(i.Gui, ImportImagePanel, x, y, w, h, NewImportImageItems(x, y, w, h), make(map[string]interface{}))
	return nil
}

func (i ImageList) LoadImage(g *gocui.Gui, v *gocui.View) error {
	i.NextPanel = ImageListPanel

	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	NewInput(i.Gui, LoadImagePanel, x, y, w, h, NewLoadImageItems(x, y, w, h), make(map[string]interface{}))
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
