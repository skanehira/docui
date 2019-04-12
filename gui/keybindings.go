package gui

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
)

func (g *Gui) setGlobalKeybinding(event *tcell.EventKey) {
	switch event.Rune() {
	case 'h':
		g.prevPanel()
	case 'l':
		g.nextPanel()
	}

	switch event.Key() {
	case tcell.KeyTab:
		g.nextPanel()
	case tcell.KeyBacktab:
		g.prevPanel()
	case tcell.KeyRight:
		g.nextPanel()
	case tcell.KeyLeft:
		g.prevPanel()
	}
}

func (g *Gui) nextPanel() {
	idx := (g.state.panels.currentPanel + 1) % len(g.state.panels.panel)
	g.switchPanel(g.state.panels.panel[idx].name())
}

func (g *Gui) prevPanel() {
	g.state.panels.currentPanel--

	if g.state.panels.currentPanel < 0 {
		g.state.panels.currentPanel = len(g.state.panels.panel) - 1
	}

	idx := (g.state.panels.currentPanel) % len(g.state.panels.panel)
	g.switchPanel(g.state.panels.panel[idx].name())
}

func (g *Gui) switchPanel(panelName string) {
	for i, panel := range g.state.panels.panel {
		if panel.name() == panelName {
			panel.focus(g)
			g.state.panels.currentPanel = i
		} else {
			panel.unfocus()
		}
	}
}

func (g *Gui) modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewGrid().
		SetColumns(0, width, 0).
		SetRows(0, height, 0).
		AddItem(p, 1, 1, 1, 1, 0, 0, true)
}

func (g *Gui) createContainerForm() {
	selectedImage := g.selectedImage()
	if selectedImage == nil {
		common.Logger.Error("please input image")
		return
	}

	image := fmt.Sprintf("%s:%s", selectedImage.Repo, selectedImage.Tag)

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("create container")
	form.SetTitleAlign(tview.AlignLeft)
	form.AddInputField("Name", "", 70, nil, nil).
		AddInputField("HostIP", "", 70, nil, nil).
		AddInputField("HostPort", "", 70, nil, nil).
		AddInputField("Port", "", 70, nil, nil).
		AddDropDown("VolumeType", []string{"bind", "volume"}, 0, func(option string, optionIndex int) {}).
		AddInputField("HostVolume", "", 70, nil, nil).
		AddInputField("Volume", "", 70, nil, nil).
		AddInputField("Image", image, 70, nil, nil).
		AddInputField("User", "", 70, nil, nil).
		AddCheckbox("Attach", false, nil).
		AddInputField("Env", "", 70, nil, nil).
		AddInputField("Cmd", "", 70, nil, nil).
		AddButton("Save", func() {
			g.createContainer(form, image)
		}).
		AddButton("Cancel", func() {
			g.pages.RemovePage("form")
			g.switchPanel("images")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 29), true).ShowPage("main")
}

func (g *Gui) createContainer(form *tview.Form, image string) {
	g.startTask("create container "+image, func(ctx context.Context) error {
		inputLabels := []string{
			"Name",
			"HostIP",
			"Port",
			"HostVolume",
			"Volume",
			"Image",
			"User",
		}

		var data = make(map[string]string)

		for _, label := range inputLabels {
			data[label] = form.GetFormItemByLabel(label).(*tview.InputField).GetText()
		}

		_, volumeType := form.GetFormItemByLabel("VolumeType").(*tview.DropDown).
			GetCurrentOption()
		data["VolymeType"] = volumeType

		isAttach := form.GetFormItemByLabel("Attach").(*tview.Checkbox).IsChecked()

		options, err := docker.Client.NewContainerOptions(data, isAttach)
		if err != nil {
			common.Logger.Errorf("cannot create container %s", err)
			return err
		}

		err = docker.Client.CreateContainer(options)
		if err != nil {
			common.Logger.Errorf("cannot create container %s", err)
			return err
		}

		g.pages.RemovePage("form")
		g.switchPanel("images")
		g.app.QueueUpdateDraw(func() {
			g.containerPanel().setEntries(g)
		})

		return nil
	})
}

func (g *Gui) pullImageForm() {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitleAlign(tview.AlignLeft)
	form.SetTitle("pull image")
	form.AddInputField("image", "", 70, nil, nil).
		AddButton("Pull", func() {
			image := form.GetFormItemByLabel("image").(*tview.InputField).GetText()
			g.pullImage(image)
		}).
		AddButton("Cancel", func() {
			g.pages.RemovePage("form")
			g.switchPanel("images")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 7), true).ShowPage("main")
}

func (g *Gui) pullImage(image string) {
	g.startTask("pull image "+image, func(ctx context.Context) error {
		g.pages.RemovePage("form")
		g.switchPanel("images")
		err := docker.Client.PullImage(image)
		if err != nil {
			common.Logger.Errorf("cannot create container %s", err)
			return err
		}

		g.app.QueueUpdateDraw(func() {
			g.imagePanel().setEntries(g)
		})

		return nil
	})
}

func (g *Gui) displayInspect(data, page string) {
	text := tview.NewTextView()
	text.SetTitle("detail").SetTitleAlign(tview.AlignLeft)
	text.SetBorder(true)
	text.SetText(data)

	text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
			g.pages.RemovePage("detail").ShowPage("main")
			g.switchPanel(page)
		}
		return event
	})

	g.pages.AddAndSwitchToPage("detail", text, true)
}

func (g *Gui) inspectImage() {
	image := g.selectedImage()

	inspect, err := docker.Client.InspectImage(image.ID)
	if err != nil {
		common.Logger.Errorf("cannot inspect image %s", err)
		return
	}

	g.displayInspect(common.StructToJSON(inspect), "images")
}

func (g *Gui) inspectContainer() {
	container := g.selectedContainer()

	inspect, err := docker.Client.InspectContainer(container.ID)
	if err != nil {
		common.Logger.Errorf("cannot inspect container %s", err)
		return
	}

	g.displayInspect(common.StructToJSON(inspect), "containers")
}

func (g *Gui) inspectVolume() {
	volume := g.selectedVolume()

	inspect, err := docker.Client.InspectVolume(volume.Name)
	if err != nil {
		common.Logger.Errorf("cannot inspect volume %s", err)
		return
	}

	g.displayInspect(common.StructToJSON(inspect), "volumes")
}

func (g *Gui) inspectNetwork() {
	network := g.selectedNetwork()

	inspect, err := docker.Client.InspectNetwork(network.ID)
	if err != nil {
		common.Logger.Errorf("cannot inspect network %s", err)
		return
	}

	g.displayInspect(common.StructToJSON(inspect), "networks")
}
