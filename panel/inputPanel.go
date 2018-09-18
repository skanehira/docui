package panel

import (
	"fmt"
	"os"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/jroimartin/gocui"
	cuidocker "github.com/skanehira/docui/docker"
)

type Position struct {
	x, y int
	w, h int
}

var activeInput = 0

type Input struct {
	*Gui
	name string
	Position
	Items
	Data map[string]interface{}
}

type Item struct {
	Label map[string]Position
	Input map[string]Position
}

type Items []Item

func NewInput(gui *Gui, name string, x, y, w, h int, items Items, data map[string]interface{}) Input {
	return Input{
		Gui:      gui,
		name:     name,
		Position: Position{x, y, w, h},
		Items:    items,
		Data:     data,
	}
}

func (i Input) Name() string {
	return i.name
}

func (i Input) SetView(g *gocui.Gui) (*gocui.View, error) {
	// create container panel
	v, err := g.SetView(i.Name(), i.x, i.y, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}

		v.Title = v.Name()
		v.Autoscroll = true
		v.Wrap = true
	}

	// create input panels
	for index, item := range i.Items {
		for name, p := range item.Label {
			if v, err := g.SetView(name, i.x+p.x, i.y+p.y, i.x+p.w, i.y+p.h); err != nil {
				if err != gocui.ErrUnknownView {
					return nil, err
				}
				v.Wrap = true
				v.Frame = false
				fmt.Fprint(v, name)
			}
		}

		for name, p := range item.Input {
			if v, err := g.SetView(name, i.x+p.x, i.y+p.y, i.x+p.w, i.y+p.h); err != nil {
				if err != gocui.ErrUnknownView {
					return nil, err
				}
				v.Wrap = true
				v.Editable = true
				v.Editor = i

				if index == 0 {
					SetCurrentPanel(g, name)
				}

				if name == "ImageInput" {
					fmt.Fprint(v, i.Data["Image"])
				}

				if name == "ContainerInput" {
					fmt.Fprint(v, i.Data["Container"])
				}

				i.SetKeyBindWithItem(name)
			}
		}
	}

	i.SetKeybinds(i.Name())
	return v, nil
}

func (i Input) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {

	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}
}

func (i Input) Init(g *Gui) {
	v, _ := i.View(i.Name())
	i.ClosePanel(g.Gui, v)
	_, err := i.SetView(g.Gui)

	if err != nil {
		panic(err)
	}
}

func (i Input) SetKeyBindWithItem(name string) {
	if err := i.SetKeybinding(name, gocui.KeyCtrlJ, gocui.ModNone, i.NextItem); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(name, gocui.KeyCtrlK, gocui.ModNone, i.PreItem); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(name, gocui.KeyCtrlW, gocui.ModNone, i.ClosePanel); err != nil {
		panic(err)
	}
	if err := i.SetKeybinding(name, gocui.KeyCtrlQ, gocui.ModNone, i.quit); err != nil {
		panic(err)
	}

	switch i.Name() {
	case CreateContainerPanel:
		if err := i.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, i.CreateContainer); err != nil {
			panic(err)
		}
	case PullImagePanel:
		if err := i.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, i.PullImage); err != nil {
			panic(err)
		}
	case ExportImagePanel:
		if err := i.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, i.ExportImage); err != nil {
			panic(err)
		}
	case ImportImagePanel:
		if err := i.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, i.ImportImage); err != nil {
			panic(err)
		}
	case LoadImagePanel:
		if err := i.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, i.LoadImage); err != nil {
			panic(err)
		}
	case ExportContainerPanel:
		if err := i.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, i.ExportContainer); err != nil {
			panic(err)
		}
	case CommitContainerPanel:
		if err := i.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, i.CommitContainer); err != nil {
			panic(err)
		}
	}
}

func (i Input) ClosePanel(g *gocui.Gui, v *gocui.View) error {
	// パネルの位置とindexをリセットしないと、再度パネルを呼び出す時にinputの移動が変になる
	// 理由不明なので時間ある時調査
	activeInput = 0
	SetCurrentPanel(g, GetKeyFromMap(i.Items[0].Input))

	for _, item := range i.Items {
		i.DeleteView(GetKeyFromMap(item.Label))
		name := GetKeyFromMap(item.Input)
		i.DeleteView(name)
		i.DeleteKeybindings(name)
	}

	i.DeleteView(i.Name())

	if i.PrePanel == "" {
		i.PrePanel = ImageListPanel
	}
	SetCurrentPanel(g, i.PrePanel)

	return nil
}

func (i Input) NextItem(g *gocui.Gui, v *gocui.View) error {

	nextIndex := (activeInput + 1) % len(i.Items)
	item := i.Items[nextIndex]

	name := GetKeyFromMap(item.Input)

	if _, err := SetCurrentPanel(g, name); err != nil {
		return err
	}

	activeInput = nextIndex
	return nil
}

func (i Input) PreItem(g *gocui.Gui, v *gocui.View) error {
	nextIndex := activeInput - 1
	if nextIndex < 0 {
		nextIndex = len(i.Items) - 1
	} else {
		nextIndex = (activeInput - 1) % len(i.Items)
	}

	item := i.Items[nextIndex]

	name := GetKeyFromMap(item.Input)

	if _, err := SetCurrentPanel(g, name); err != nil {
		return err
	}

	activeInput = nextIndex
	return nil
}

func (i Input) CreateContainer(g *gocui.Gui, v *gocui.View) error {
	defer i.ClosePanel(g, v)
	defer i.RefreshAllPanel()
	i.PrePanel = ImageListPanel

	data := make(map[string]string)
	for _, item := range i.Items {
		name := GetKeyFromMap(item.Label)

		v, err := i.View(GetKeyFromMap(item.Input))

		if err != nil {
			return err
		}

		data[name] = ReadLine(v, nil)
	}

	options := cuidocker.NewContainerOptions(data)

	if err := i.Docker.CreateContainerWithOptions(options); err != nil {
		i.DispMessage(err.Error(), ImageListPanel)
		return nil
	}

	return nil
}

func (i Input) ExportContainer(g *gocui.Gui, v *gocui.View) error {
	defer i.ClosePanel(g, v)
	defer i.RefreshAllPanel()
	i.PrePanel = ContainerListPanel

	path := ReadLine(v, nil)

	if path == "" {
		return nil
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	options := docker.ExportContainerOptions{
		ID:           i.Data["ID"].(string),
		OutputStream: file,
	}

	if err := i.Docker.ExportContainerWithOptions(options); err != nil {
		i.DispMessage(err.Error(), ContainerListPanel)
		return nil
	}

	return nil
}

func (i Input) CommitContainer(g *gocui.Gui, v *gocui.View) error {
	i.PrePanel = ContainerListPanel
	defer i.ClosePanel(g, v)
	defer i.RefreshAllPanel()

	config := make(map[string]string)
	for _, item := range i.Items {
		name := GetKeyFromMap(item.Label)

		v, err := i.View(GetKeyFromMap(item.Input))

		if err != nil {
			return err
		}

		value := ReadLine(v, nil)
		if name == "Tag" && value == "" {
			value = "latest"
		}

		config[name] = value
	}

	options := docker.CommitContainerOptions{
		Container:  i.Data["Container"].(string),
		Repository: config["Repository"],
		Tag:        config["Tag"],
	}

	if err := i.Docker.CommitContainerWithOptions(options); err != nil {
		i.DispMessage(err.Error(), ImageListPanel)
		return nil
	}

	return nil
}

func (i Input) PullImage(g *gocui.Gui, v *gocui.View) error {
	defer i.ClosePanel(g, v)
	defer i.RefreshAllPanel()

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

	options := docker.PullImageOptions{
		Repository: name,
		Tag:        tag,
	}

	if err := i.Docker.PullImageWithOptions(options); err != nil {
		i.DispMessage(err.Error(), ImageListPanel)
		return nil
	}

	i.PrePanel = ImageListPanel

	return nil
}

func (i Input) ExportImage(g *gocui.Gui, v *gocui.View) error {
	defer i.ClosePanel(g, v)
	defer i.RefreshAllPanel()

	path := ReadLine(v, nil)
	id := i.Data["ID"].(string)

	if path == "" {
		return nil
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	options := docker.ExportImageOptions{
		Name:         id,
		OutputStream: file,
	}

	if err := i.Docker.SaveImageWithOptions(options); err != nil {
		i.DispMessage(err.Error(), ImageListPanel)
		return nil
	}

	i.PrePanel = ImageListPanel

	return nil
}

func (i Input) ImportImage(g *gocui.Gui, v *gocui.View) error {
	defer i.ClosePanel(g, v)
	defer i.RefreshAllPanel()

	data := make(map[string]string)
	for _, item := range i.Items {
		name := GetKeyFromMap(item.Label)

		v, err := i.View(GetKeyFromMap(item.Input))

		if err != nil {
			return err
		}

		value := ReadLine(v, nil)

		if value == "" {
			if name == "tag" {
				value = "latest"
				continue
			}
			return nil
		}
		data[name] = value
	}

	options := docker.ImportImageOptions{
		Repository: data["repository"],
		Source:     data["path"],
		Tag:        data["tag"],
	}

	if err := i.Docker.ImportImageWithOptions(options); err != nil {
		i.DispMessage(err.Error(), ImageListPanel)
		return nil
	}

	i.PrePanel = ImageListPanel

	return nil
}

func (i Input) LoadImage(g *gocui.Gui, v *gocui.View) error {
	path := ReadLine(v, nil)
	if path == "" {
		return nil
	}

	i.ClosePanel(g, v)

	if err := i.Docker.LoadImageWithPath(path); err != nil {
		i.DispMessage(err.Error(), ImageListPanel)
		return nil
	}

	i.PrePanel = ImageListPanel
	defer i.RefreshAllPanel()

	return nil
}

func (i Input) RefreshPanel(g *gocui.Gui, v *gocui.View) error {
	i.Init(i.Gui)
	return nil
}

func GetKeyFromMap(m map[string]Position) string {
	var key string
	for k, _ := range m {
		key = k
	}

	return key
}
