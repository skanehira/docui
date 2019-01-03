package panel

import (
	"fmt"

	"github.com/jroimartin/gocui"
)

type Navigate struct {
	*Gui
	name string
	Position
	Navi map[string]string
}

func NewNavigate(g *Gui, name string, x, y, w, h int) Navigate {
	return Navigate{
		Gui:      g,
		name:     name,
		Position: Position{x, y, w, h},
		Navi:     newNavi(),
	}
}

func (n Navigate) Name() string {
	return n.name
}

func (n Navigate) SetView(g *gocui.Gui) error {
	v, err := g.SetView(n.name, n.x, n.y, n.w, n.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Frame = false
		v.FgColor = gocui.ColorYellow
	}

	n.Refresh(g, v)
	return nil
}

func (n Navigate) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
}

func (n Navigate) Refresh(g *gocui.Gui, v *gocui.View) error {
	n.Update(func(g *gocui.Gui) error {
		currentView := g.CurrentView().Name()

		n.SetNavigate(currentView)
		return nil
	})

	return nil
}

func (n Navigate) SetNavigate(name string) *gocui.View {
	v, err := n.View(n.name)
	if err != nil {
		panic(err)
	}
	v.Clear()

	fmt.Fprint(v, n.Navi[name])
	return v
}

func newNavi() map[string]string {
	return map[string]string{
		ImageListPanel:         "j/k: select image, p: pull image, i: import image, s: save image\nCtrl+l: load image, ctrl+f: search image, d: remove image, Ctrl+d: remove dagling images, c: create container, Enter/o: inspect image, Ctrl+r: refresh images iist",
		PullImagePanel:         "Esc/Ctrl+w: close panel, Enter: pull image",
		ContainerListPanel:     "j/k: select container, e: export container, c: commit container\nu: start container, s: stop container, d: remove container, Enter/o: inspect container, Ctrl+r: refresh container list",
		DetailPanel:            "j/k: cursor down/up, d/u: page down/up",
		CreateContainerPanel:   "Ctrl+j/k: change input, Esc/Ctrl+w: close panel, Enter: create container",
		SaveImagePanel:         "Esc/Ctrl+w: close panel, Enter: save image",
		ImportImagePanel:       "Esc/Ctrl+w: close panel, Enter: import image",
		LoadImagePanel:         "Esc/Ctrl+w: close panel, Enter: load image",
		ExportContainerPanel:   "Esc/Ctrl+w: close panel, Enter: export container",
		CommitContainerPanel:   "Ctrl+j/k: change input, Esc/Ctrl+w: close panel, Enter: commit container",
		SearchImagePanel:       "Esc/Ctrl+w: close panel, Enter: serach image",
		SearchImageResultPanel: "j/k: select image, Esc/Ctrl+w: close panel, Enter: pull image",
		VolumeListPanel:        "j/k: select volume, c: create volume, d: remove volume, p: prune volumes, Enter/o: inspect volume, Ctrl+r: refresh volume list",
		CreateVolumePanel:      "Esc/Ctrl+w: close panel, Enter: create volume",
		NetworkListPanel:       "j/k: cursor down/up, d: remove network, o/Enter: inspect network",
	}

}
