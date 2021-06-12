package gui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gdamore/tcell/v2"
	"github.com/pkg/browser"
	"github.com/rivo/tview"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker"
)

var inputWidth = 70

func (g *Gui) setGlobalKeybinding(event *tcell.EventKey) {
	switch event.Rune() {
	case 'h':
		g.prevPanel()
	case 'l':
		g.nextPanel()
	case 'q':
		g.Stop()
	case '/':
		g.filter()
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

func (g *Gui) filter() {
	currentPanel := g.state.panels.panel[g.state.panels.currentPanel]
	if currentPanel.name() == "tasks" {
		return
	}
	currentPanel.setFilterWord("")
	currentPanel.updateEntries(g)

	viewName := "filter"
	searchInput := tview.NewInputField().SetLabel("Word")
	searchInput.SetLabelWidth(6)
	searchInput.SetTitle("filter")
	searchInput.SetTitleAlign(tview.AlignLeft)
	searchInput.SetBorder(true)

	closeSearchInput := func() {
		g.closeAndSwitchPanel(viewName, g.state.panels.panel[g.state.panels.currentPanel].name())
	}

	searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			closeSearchInput()
		}
	})

	searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			closeSearchInput()
		}
		return event
	})

	searchInput.SetChangedFunc(func(text string) {
		currentPanel.setFilterWord(text)
		currentPanel.updateEntries(g)
	})

	g.pages.AddAndSwitchToPage(viewName, g.modal(searchInput, 80, 3), true).ShowPage("main")
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

func (g *Gui) createContainerForm() {
	selectedImage := g.selectedImage()
	if selectedImage == nil {
		common.Logger.Error("please input image")
		return
	}

	image := fmt.Sprintf("%s:%s", selectedImage.Repo, selectedImage.Tag)

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle("Create container")
	form.SetTitleAlign(tview.AlignLeft)

	form.AddInputField("Name", "", inputWidth, nil, nil).
		AddInputField("HostIP", "", inputWidth, nil, nil).
		AddInputField("HostPort", "", inputWidth, nil, nil).
		AddInputField("Port", "", inputWidth, nil, nil).
		AddDropDown("VolumeType", []string{"bind", "volume"}, 0, func(option string, optionIndex int) {}).
		AddInputField("HostVolume", "", inputWidth, nil, nil).
		AddInputField("Volume", "", inputWidth, nil, nil).
		AddInputField("Image", image, inputWidth, nil, nil).
		AddInputField("User", "", inputWidth, nil, nil).
		AddCheckbox("Attach", false, nil).
		AddInputField("Env", "", inputWidth, nil, nil).
		AddInputField("Cmd", "", inputWidth, nil, nil).
		AddButton("Create", func() {
			g.createContainer(form, image)
		}).
		AddButton("Cancel", func() {
			g.closeAndSwitchPanel("form", "images")
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
		data["VolumeType"] = volumeType

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

		g.closeAndSwitchPanel("form", "images")
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
	form.SetTitle("Pull image")
	form.AddInputField("Image", "", inputWidth, nil, nil).
		AddButton("Pull", func() {
			image := form.GetFormItemByLabel("Image").(*tview.InputField).GetText()
			g.pullImage(image, "form", "images")
		}).
		AddButton("Cancel", func() {
			g.closeAndSwitchPanel("form", "images")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 7), true).ShowPage("main")
}

func (g *Gui) pullImage(image, closePanel, switchPanel string) {
	g.startTask("Pull image "+image, func(ctx context.Context) error {
		g.closeAndSwitchPanel(closePanel, switchPanel)
		err := docker.Client.PullImage(image)
		if err != nil {
			common.Logger.Errorf("cannot pull an image %s", err)
			return err
		}

		g.imagePanel().updateEntries(g)

		return nil
	})
}

func (g *Gui) displayInspect(data, page string) {
	text := tview.NewTextView()
	text.SetTitle("Detail").SetTitleAlign(tview.AlignLeft)
	text.SetBorder(true)
	text.SetText(data)

	text.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
			g.closeAndSwitchPanel("detail", page)
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

func (g *Gui) renameContainerForm() {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitleAlign(tview.AlignLeft)
	form.SetTitle("Rename container")
	form.AddInputField("NewName", "", inputWidth, nil, nil).
		AddButton("Rename", func() {
			image := form.GetFormItemByLabel("NewName").(*tview.InputField).GetText()
			g.renameContainer(image, "form", "containers")
		}).
		AddButton("Cancel", func() {
			g.closeAndSwitchPanel("form", "containers")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 7), true).ShowPage("main")
}

func (g *Gui) renameContainer(newName, closePanel, switchPanel string) {
	g.startTask("Renaming container "+newName, func(ctx context.Context) error {
		g.closeAndSwitchPanel(closePanel, switchPanel)
		oldContainer := g.selectedContainer()
		if oldContainer == nil {
			err := errors.New("specified container is nil")
			common.Logger.Errorf("cannot rename container %s", err)
			return err
		}

		err := docker.Client.RenameContainer(oldContainer.ID, newName)
		if err != nil {
			common.Logger.Errorf("cannot create container %s", err)
			return err
		}

		g.containerPanel().updateEntries(g)

		return nil
	})
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

func (g *Gui) removeImage() {
	image := g.selectedImage()

	g.confirm("Do you want to remove the image?", "Done", "images", func() {
		g.startTask(fmt.Sprintf("remove image %s:%s", image.Repo, image.Tag), func(ctx context.Context) error {
			if err := docker.Client.RemoveImage(image.ID); err != nil {
				common.Logger.Errorf("cannot remove the image %s", err)
				return err
			}
			g.imagePanel().updateEntries(g)
			return nil
		})
	})
}

func (g *Gui) removeContainer() {
	container := g.selectedContainer()

	g.confirm("Do you want to remove the container?", "Done", "containers", func() {
		g.startTask(fmt.Sprintf("remove container %s", container.Name), func(ctx context.Context) error {
			if err := docker.Client.RemoveContainer(container.ID); err != nil {
				common.Logger.Errorf("cannot remove the container %s", err)
				return err
			}
			g.containerPanel().updateEntries(g)
			return nil
		})
	})
}

func (g *Gui) removeVolume() {
	volume := g.selectedVolume()

	g.confirm("Do you want to remove the volume?", "Done", "volumes", func() {
		g.startTask(fmt.Sprintf("remove volume %s", volume.Name), func(ctx context.Context) error {
			if err := docker.Client.RemoveVolume(volume.Name); err != nil {
				common.Logger.Errorf("cannot remove the volume %s", err)
				return err
			}
			g.volumePanel().updateEntries(g)
			return nil
		})
	})
}

func (g *Gui) removeNetwork() {
	network := g.selectedNetwork()

	g.confirm("Do you want to remove the network?", "Done", "networks", func() {
		g.startTask(fmt.Sprintf("remove network %s", network.Name), func(ctx context.Context) error {
			if err := docker.Client.RemoveNetwork(network.ID); err != nil {
				common.Logger.Errorf("cannot remove the network %s", err)
				return err
			}
			g.networkPanel().updateEntries(g)
			return nil
		})
	})
}

func (g *Gui) startContainer() {
	container := g.selectedContainer()

	g.startTask(fmt.Sprintf("start container %s", container.Name), func(ctx context.Context) error {
		if err := docker.Client.StartContainer(container.ID); err != nil {
			common.Logger.Errorf("cannot start container %s", err)
			return err
		}

		g.containerPanel().updateEntries(g)
		return nil
	})
}

func (g *Gui) stopContainer() {
	container := g.selectedContainer()

	g.startTask(fmt.Sprintf("stop container %s", container.Name), func(ctx context.Context) error {

		if err := docker.Client.StopContainer(container.ID); err != nil {
			common.Logger.Errorf("cannot stop container %s", err)
			return err
		}

		g.containerPanel().updateEntries(g)
		return nil
	})
}

func (g *Gui) exportContainerForm() {
	inputWidth := 70

	container := g.selectedContainer()
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitleAlign(tview.AlignLeft)
	form.SetTitle("Export container")
	form.AddInputField("Path", "", inputWidth, nil, nil).
		AddInputField("Container", container.Name, inputWidth, nil, nil).
		AddButton("Create", func() {
			path := form.GetFormItemByLabel("Path").(*tview.InputField).GetText()
			container := form.GetFormItemByLabel("Container").(*tview.InputField).GetText()

			g.exportContainer(path, container)
		}).
		AddButton("Cancel", func() {
			g.closeAndSwitchPanel("form", "containers")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 9), true).ShowPage("main")
}

func (g *Gui) exportContainer(path, container string) {
	g.startTask("export container "+container, func(ctx context.Context) error {
		g.closeAndSwitchPanel("form", "containers")
		err := docker.Client.ExportContainer(container, path)
		if err != nil {
			common.Logger.Errorf("cannot export container %s", err)
			return err
		}

		return nil
	})
}

func (g *Gui) loadImageForm() {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitleAlign(tview.AlignLeft)
	form.SetTitle("Load image")
	form.AddInputField("Path", "", inputWidth, nil, nil).
		AddButton("Load", func() {
			path := form.GetFormItemByLabel("Path").(*tview.InputField).GetText()
			g.loadImage(path)
		}).
		AddButton("Cancel", func() {
			g.closeAndSwitchPanel("form", "images")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 7), true).ShowPage("main")
}

func (g *Gui) loadImage(path string) {
	g.startTask("load image "+filepath.Base(path), func(ctx context.Context) error {
		g.closeAndSwitchPanel("form", "images")
		if err := docker.Client.LoadImage(path); err != nil {
			common.Logger.Errorf("cannot load image %s", err)
			return err
		}

		g.imagePanel().updateEntries(g)
		return nil
	})
}

func (g *Gui) importImageForm() {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitleAlign(tview.AlignLeft)
	form.SetTitle("Import image")
	form.AddInputField("Repository", "", inputWidth, nil, nil).
		AddInputField("Tag", "", inputWidth, nil, nil).
		AddInputField("Path", "", inputWidth, nil, nil).
		AddButton("Load", func() {
			repository := form.GetFormItemByLabel("Repository").(*tview.InputField).GetText()
			tag := form.GetFormItemByLabel("Tag").(*tview.InputField).GetText()
			path := form.GetFormItemByLabel("Path").(*tview.InputField).GetText()
			g.importImage(path, repository, tag)
		}).
		AddButton("Cancel", func() {
			g.closeAndSwitchPanel("form", "images")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 11), true).ShowPage("main")
}

func (g *Gui) importImage(file, repo, tag string) {
	g.startTask("import image "+file, func(ctx context.Context) error {
		g.closeAndSwitchPanel("form", "images")

		if err := docker.Client.ImportImage(repo, tag, file); err != nil {
			common.Logger.Errorf("cannot load image %s", err)
			return err
		}

		g.imagePanel().updateEntries(g)
		return nil
	})
}

func (g *Gui) saveImageForm() {
	image := g.selectedImage()
	imageName := fmt.Sprintf("%s:%s", image.Repo, image.Tag)

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitleAlign(tview.AlignLeft)
	form.SetTitle("Save image")
	form.AddInputField("Path", "", inputWidth, nil, nil).
		AddInputField("Image", imageName, inputWidth, nil, nil).
		AddButton("Save", func() {
			image := form.GetFormItemByLabel("Image").(*tview.InputField).GetText()
			path := form.GetFormItemByLabel("Path").(*tview.InputField).GetText()
			g.saveImage(image, path)
		}).
		AddButton("Cancel", func() {
			g.closeAndSwitchPanel("form", "images")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 9), true).ShowPage("main")

}

func (g *Gui) saveImage(image, path string) {
	g.startTask("save image "+image, func(ctx context.Context) error {
		g.closeAndSwitchPanel("form", "images")

		if err := docker.Client.SaveImage([]string{image}, path); err != nil {
			common.Logger.Errorf("cannot save image %s", err)
			return err
		}
		return nil
	})

}

func (g *Gui) commitContainerForm() {
	container := g.selectedContainer()

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitleAlign(tview.AlignLeft)
	form.SetTitle("Commit container")
	form.AddInputField("Repository", "", inputWidth, nil, nil).
		AddInputField("Tag", "", inputWidth, nil, nil).
		AddInputField("Container", container.Name, inputWidth, nil, nil).
		AddButton("Commit", func() {
			repo := form.GetFormItemByLabel("Repository").(*tview.InputField).GetText()
			tag := form.GetFormItemByLabel("Tag").(*tview.InputField).GetText()
			con := form.GetFormItemByLabel("Container").(*tview.InputField).GetText()
			g.commitContainer(repo, tag, con)
		}).
		AddButton("Cancel", func() {
			g.closeAndSwitchPanel("form", "containers")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 11), true).ShowPage("main")
}

func (g *Gui) openBrowser() {
	container := g.selectedContainer()
	if container.WebPort == "" {
		return
	}
	link := "http://" + container.WebPort
	browser.OpenURL(link)
}

func (g *Gui) commitContainer(repo, tag, container string) {
	g.startTask("commit container "+container, func(ctx context.Context) error {
		g.closeAndSwitchPanel("form", "containers")

		if err := docker.Client.CommitContainer(container, types.ContainerCommitOptions{Reference: repo + ":" + tag}); err != nil {
			common.Logger.Errorf("cannot commit container %s", err)
			return err
		}

		g.imagePanel().updateEntries(g)
		return nil
	})
}

func (g *Gui) attachContainerForm() {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitleAlign(tview.AlignLeft)
	form.SetTitle("Exec container")
	form.AddInputField("Cmd", "", inputWidth, nil, nil).
		AddButton("Exec", func() {
			cmd := form.GetFormItemByLabel("Cmd").(*tview.InputField).GetText()
			g.attachContainer(g.selectedContainer().ID, cmd)
		}).
		AddButton("Cancel", func() {
			g.closeAndSwitchPanel("form", "containers")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 7), true).ShowPage("main")
}

func (g *Gui) attachContainer(container, cmd string) {
	g.closeAndSwitchPanel("form", "containers")

	if !g.app.Suspend(func() {
		g.stopMonitoring()
		if err := docker.Client.AttachExecContainer(container, cmd); err != nil {
			common.Logger.Errorf("cannot attach container %s", err)
		}

		g.startMonitoring()
	}) {
		common.Logger.Error("cannot suspend tview")
	}
}

func (g *Gui) createVolumeForm() {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitleAlign(tview.AlignLeft)
	form.SetTitle("Create volume")
	form.AddInputField("Name", "", inputWidth, nil, nil).
		AddInputField("Labels", "", inputWidth, nil, nil).
		AddInputField("Driver", "", inputWidth, nil, nil).
		AddInputField("Options", "", inputWidth, nil, nil).
		AddButton("Create", func() {
			g.createVolume(form)
		}).
		AddButton("Cancel", func() {
			g.closeAndSwitchPanel("form", "volumes")
		})

	g.pages.AddAndSwitchToPage("form", g.modal(form, 80, 13), true).ShowPage("main")
}

func (g *Gui) createVolume(form *tview.Form) {
	var data = make(map[string]string)
	inputLabels := []string{
		"Name",
		"Labels",
		"Driver",
		"Options",
	}

	for _, label := range inputLabels {
		data[label] = form.GetFormItemByLabel(label).(*tview.InputField).GetText()
	}

	g.startTask("create volume "+data["Name"], func(ctx context.Context) error {
		options := docker.Client.NewCreateVolumeOptions(data)

		if err := docker.Client.CreateVolume(options); err != nil {
			common.Logger.Errorf("cannot create volume %s", err)
			return err
		}

		g.closeAndSwitchPanel("form", "volumes")
		g.app.QueueUpdateDraw(func() {
			g.volumePanel().setEntries(g)
		})

		return nil
	})
}

func (g *Gui) tailContainerLog() {
	container := g.selectedContainer()
	if container == nil {
		common.Logger.Errorf("cannot start tail container: selected container is null")
		return
	}

	if !g.app.Suspend(func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		errCh := make(chan error)

		var reader io.ReadCloser
		var err error

		go func() {
			reader, err = docker.Client.ContainerLogStream(container.ID)
			if err != nil {
				common.Logger.Error(err)
				errCh <- err
			}
			defer reader.Close()

			_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, reader)
			if err != nil {
				common.Logger.Error(err)
				errCh <- err
			}
			return
		}()

		select {
		case err := <-errCh:
			common.Logger.Error(err)
			reader.Close()
			return
		case <-sigint:
			reader.Close()
			return
		}
	}) {
		common.Logger.Error("cannot suspend tview")
	}
}

func (g *Gui) killContainer() {
	container := g.selectedContainer()
	if container == nil {
		common.Logger.Errorf("cannot kill container: selected container is null")
		return
	}

	g.confirm("Do you want to kill the container?", "Done", "containers", func() {
		g.startTask(fmt.Sprintf("kill container %s", container.Name), func(ctx context.Context) error {
			if err := docker.Client.KillContainer(container.ID); err != nil {
				common.Logger.Errorf("cannot kill the container %s", err)
				return err
			}
			g.containerPanel().updateEntries(g)
			return nil
		})
	})
}
