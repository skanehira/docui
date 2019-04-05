package gui

import (
	"github.com/docker/docker/api/types"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
)

// image image info.
type image struct {
	ID      string
	Repo    string
	Tag     string
	Created string
	Size    string
}

type images struct {
	*tview.Table
	images []*image
}

func newImages() *images {
	images := &images{
		Table: tview.NewTable().SetSelectable(true, false),
	}

	images.SetTitle("image list").SetTitleAlign(tview.AlignLeft)
	images.SetBorder(true)

	images.getEntries()
	images.setEntries()
	return images
}

func (i *images) getSelected() interface{} {
	row, _ := i.GetSelection()
	if len(i.images) == 0 {
		return nil
	}
	return i.images[row-1]
}

func (i *images) getEntries() {
	images, err := docker.Client.Images(types.ImageListOptions{})
	if err != nil {
		return
	}

	for _, imgInfo := range images {
		for _, repoTag := range imgInfo.RepoTags {
			repo, tag := common.ParseRepoTag(repoTag)

			i.images = append(i.images, &image{
				ID:      imgInfo.ID[7:19],
				Repo:    repo,
				Tag:     tag,
				Created: common.ParseDateToString(imgInfo.Created),
				Size:    common.ParseSizeToString(imgInfo.Size),
			})
		}
	}
}

func (i *images) setEntries() {
	table := i.Clear().Select(0, 0).SetFixed(1, 1).SetSeparator('|')

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
			Color:           tcell.ColorYellow,
			BackgroundColor: tcell.ColorDefault,
		})
	}

	for i, image := range i.images {
		table.SetCell(i+1, 0, tview.NewTableCell(image.ID).
			SetTextColor(tcell.ColorDarkCyan).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 1, tview.NewTableCell(image.Repo).
			SetTextColor(tcell.ColorDarkCyan).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 2, tview.NewTableCell(image.Tag).
			SetTextColor(tcell.ColorDarkCyan).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 3, tview.NewTableCell(image.Created).
			SetTextColor(tcell.ColorDarkCyan).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 4, tview.NewTableCell(image.Size).
			SetTextColor(tcell.ColorDarkCyan).
			SetMaxWidth(1).
			SetExpansion(1))
	}
}
