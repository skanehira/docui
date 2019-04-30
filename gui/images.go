package gui

import (
	"github.com/docker/docker/api/types"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
)

type image struct {
	ID      string
	Repo    string
	Tag     string
	Created string
	Size    string
}

type images struct {
	*tview.Table
}

func newImages(g *Gui) *images {
	images := &images{
		Table: tview.NewTable().SetSelectable(true, false).Select(0, 0).SetFixed(1, 1),
	}

	images.SetTitle("image list").SetTitleAlign(tview.AlignLeft)
	images.SetBorder(true)
	images.setEntries(g)
	images.setKeybinding(g)
	return images
}

func (i *images) name() string {
	return "images"
}

func (i *images) setKeybinding(g *Gui) {
	i.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		g.setGlobalKeybinding(event)
		switch event.Key() {
		case tcell.KeyEnter:
			g.inspectImage()
		case tcell.KeyCtrlL:
			g.loadImageForm()
		}

		switch event.Rune() {
		case 'c':
			g.createContainerForm()
		case 'p':
			g.pullImageForm()
		case 'd':
			g.removeImage()
		case 'i':
			g.importImageForm()
		case 's':
			g.saveImageForm()
		case 'f':
			newSearchInputField(g)
		}

		return event
	})
}

func (i *images) entries(g *Gui) {
	images, err := docker.Client.Images(types.ImageListOptions{})
	if err != nil {
		return
	}

	g.state.resources.images = make([]*image, 0)

	for _, imgInfo := range images {
		for _, repoTag := range imgInfo.RepoTags {
			repo, tag := common.ParseRepoTag(repoTag)

			g.state.resources.images = append(g.state.resources.images, &image{
				ID:      imgInfo.ID[7:19],
				Repo:    repo,
				Tag:     tag,
				Created: common.ParseDateToString(imgInfo.Created),
				Size:    common.ParseSizeToString(imgInfo.Size),
			})
		}
	}
}

func (i *images) setEntries(g *Gui) {
	i.entries(g)
	table := i.Clear()

	headers := []string{
		"ID",
		"Repo",
		"Tag",
		"Created",
		"Size",
	}

	for i, header := range headers {
		table.SetCell(0, i, &tview.TableCell{
			Text:            header,
			NotSelectable:   true,
			Align:           tview.AlignLeft,
			Color:           tcell.ColorWhite,
			BackgroundColor: tcell.ColorDefault,
			Attributes:      tcell.AttrBold,
		})
	}

	for i, image := range g.state.resources.images {
		table.SetCell(i+1, 0, tview.NewTableCell(image.ID).
			SetTextColor(tcell.ColorLightYellow).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 1, tview.NewTableCell(image.Repo).
			SetTextColor(tcell.ColorLightYellow).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 2, tview.NewTableCell(image.Tag).
			SetTextColor(tcell.ColorLightYellow).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 3, tview.NewTableCell(image.Created).
			SetTextColor(tcell.ColorLightYellow).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 4, tview.NewTableCell(image.Size).
			SetTextColor(tcell.ColorLightYellow).
			SetMaxWidth(1).
			SetExpansion(1))
	}
}

func (i *images) updateEntries(g *Gui) {
	g.app.QueueUpdateDraw(func() {
		i.setEntries(g)
	})
}

func (i *images) focus(g *Gui) {
	i.SetSelectable(true, false)
	g.app.SetFocus(i)
}

func (i *images) unfocus() {
	i.SetSelectable(false, false)
}
