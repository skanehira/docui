package gui

import (
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
)

var replacer = strings.NewReplacer("T", " ", "Z", "")

type volume struct {
	Name       string
	MountPoint string
	Driver     string
	Created    string
}

type volumes struct {
	*tview.Table
}

func newVolumes(g *Gui) *volumes {
	volumes := &volumes{
		Table: tview.NewTable().SetSelectable(true, false),
	}

	volumes.SetTitle("volume list").SetTitleAlign(tview.AlignLeft)
	volumes.SetBorder(true)
	volumes.setEntries(g)
	return volumes
}

func (n *volumes) name() string {
	return "volumes"
}

func (n *volumes) setKeybinding(f func(event *tcell.EventKey) *tcell.EventKey) {
	n.SetInputCapture(f)
}

func (n *volumes) entries(g *Gui) {
	volumes, err := docker.Client.Volumes()
	if err != nil {
		common.Logger.Error(err)
		return
	}

	keys := make([]string, 0, len(volumes))
	tmpMap := make(map[string]*volume)

	for _, v := range volumes {
		tmpMap[v.Name] = &volume{
			Name:       v.Name,
			MountPoint: v.Mountpoint,
			Driver:     v.Driver,
			Created:    replacer.Replace(v.CreatedAt),
		}

		keys = append(keys, v.Name)
	}

	for _, key := range common.SortKeys(keys) {
		g.state.resources.volumes = append(g.state.resources.volumes, tmpMap[key])
	}
}

func (n *volumes) setEntries(g *Gui) {
	n.entries(g)
	table := n.Clear().Select(0, 0).SetFixed(1, 1)

	headers := []string{
		"Name",
		"MountPoint",
		"Driver",
		"Created",
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

	for i, network := range g.state.resources.volumes {
		table.SetCell(i+1, 0, tview.NewTableCell(network.Name).
			SetTextColor(tcell.ColorLightPink).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 1, tview.NewTableCell(network.MountPoint).
			SetTextColor(tcell.ColorLightPink).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 2, tview.NewTableCell(network.Driver).
			SetTextColor(tcell.ColorLightPink).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 3, tview.NewTableCell(network.Created).
			SetTextColor(tcell.ColorLightPink).
			SetMaxWidth(1).
			SetExpansion(1))
	}
}
