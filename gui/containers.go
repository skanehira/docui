package gui

import (
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
)

type container struct {
	ID      string
	Name    string
	Image   string
	Status  string
	Created string
	Port    string
	WebPort string
}

type containers struct {
	*tview.Table
	filterWord string
}

func newContainers(g *Gui) *containers {
	containers := &containers{
		Table: tview.NewTable().SetSelectable(true, false).Select(0, 0).SetFixed(1, 1),
	}

	containers.SetTitle("container list").SetTitleAlign(tview.AlignLeft)
	containers.SetBorder(true)
	containers.setEntries(g)
	containers.setKeybinding(g)
	return containers
}

func (c *containers) name() string {
	return "containers"
}

func (c *containers) setKeybinding(g *Gui) {
	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		g.setGlobalKeybinding(event)
		switch event.Key() {
		case tcell.KeyEnter:
			g.inspectContainer()
		case tcell.KeyCtrlE:
			g.attachContainerForm()
		case tcell.KeyCtrlL:
			g.tailContainerLog()
		case tcell.KeyCtrlK:
			g.killContainer()
		case tcell.KeyCtrlR:
			c.setEntries(g)
		}

		switch event.Rune() {
		case 'd':
			g.removeContainer()
		case 'r':
			g.renameContainerForm()
		case 'u':
			g.startContainer()
		case 's':
			g.stopContainer()
		case 'e':
			g.exportContainerForm()
		case 'c':
			g.commitContainerForm()
		case 'w':
			g.openBrowser()
		}

		return event
	})
}

func (c *containers) entries(g *Gui) {
	containers, err := docker.Client.Containers(types.ContainerListOptions{All: true})
	if err != nil {
		return
	}

	g.state.resources.containers = make([]*container, 0)

	for _, con := range containers {
		if strings.Index(con.Names[0][1:], c.filterWord) == -1 {
			continue
		}

		g.state.resources.containers = append(g.state.resources.containers, &container{
			ID:      con.ID[:12],
			Image:   con.Image,
			Name:    con.Names[0][1:],
			Status:  con.Status,
			Created: common.ParseDateToString(con.Created),
			Port:    common.ParsePortToString(con.Ports),
			WebPort: common.WebPort(con.Ports),
		})
	}
}

func (c *containers) setEntries(g *Gui) {
	c.entries(g)
	table := c.Clear()

	headers := []string{
		"ID",
		"Name",
		"Image",
		"Status",
		"Created",
		"Port",
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

	for i, container := range g.state.resources.containers {
		table.SetCell(i+1, 0, tview.NewTableCell(container.ID).
			SetTextColor(tcell.ColorLightGreen).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 1, tview.NewTableCell(container.Name).
			SetTextColor(tcell.ColorLightGreen).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 2, tview.NewTableCell(container.Image).
			SetTextColor(tcell.ColorLightGreen).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 3, tview.NewTableCell(container.Status).
			SetTextColor(tcell.ColorLightGreen).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 4, tview.NewTableCell(container.Created).
			SetTextColor(tcell.ColorLightGreen).
			SetMaxWidth(1).
			SetExpansion(1))

		table.SetCell(i+1, 5, tview.NewTableCell(container.Port).
			SetTextColor(tcell.ColorLightGreen).
			SetMaxWidth(1).
			SetExpansion(1))
	}
}

func (c *containers) focus(g *Gui) {
	c.SetSelectable(true, false)
	g.app.SetFocus(c)
}

func (c *containers) unfocus() {
	c.SetSelectable(false, false)
}

func (c *containers) updateEntries(g *Gui) {
	go g.app.QueueUpdateDraw(func() {
		c.setEntries(g)
	})
}

func (c *containers) setFilterWord(word string) {
	c.filterWord = word
}

func (c *containers) monitoringContainers(g *Gui) {
	common.Logger.Info("start monitoring containers")
	ticker := time.NewTicker(5 * time.Second)

LOOP:
	for {
		select {
		case <-ticker.C:
			c.updateEntries(g)
		case <-g.state.stopChans["container"]:
			ticker.Stop()
			break LOOP
		}
	}
	common.Logger.Info("stop monitoring containers")
}
