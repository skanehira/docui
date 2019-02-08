package panel

import (
	"fmt"
	"runtime"

	"github.com/jroimartin/gocui"
)

// Info have docui and docker info.
type Info struct {
	*Gui
	Position
	name   string
	Docker *DockerInfo
	Host   *HostInfo
	Docui  *Docui
}

// Docui docui's info.
type Docui struct {
	Name    string
	Version string
}

// DockerInfo docker's info.
type DockerInfo struct {
	HostName      string
	ServerVersion string
	APIVersion    string
	KernelVersion string
	OSType        string
	Architecture  string
	Endpoint      string
	Containers    int
	Images        int
	MemTotal      string
}

// HostInfo host os info.
type HostInfo struct {
	OSType       string
	Architecture string
}

// SetView set up info panel.
func (i *Info) SetView(g *gocui.Gui) error {
	v, err := g.SetView(i.name, i.x, i.y, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Frame = false
		v.FgColor = gocui.ColorYellow | gocui.AttrBold

		dockerAPI := fmt.Sprintf("api:%s", i.Docker.APIVersion)
		dockerVersion := fmt.Sprintf("version:%s", i.Docker.ServerVersion)
		dockerEndpoint := fmt.Sprintf("endpoint:%s", i.Docker.Endpoint)
		docuiVersion := fmt.Sprintf("version:%s", i.Docui.Version)

		// print info
		fmt.Fprintf(v, "Docker	|	%s %s %s\ndocui	 | %s", dockerAPI, dockerVersion, dockerEndpoint, docuiVersion)
	}

	return nil
}

// Name return panel name.
func (i *Info) Name() string {
	return i.name
}

// Edit do nothing
func (i *Info) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	// do nothing
}

// Refresh do nothing
func (i *Info) Refresh(g *gocui.Gui, v *gocui.View) error {
	// do nothing
	return nil
}

// NewInfo create info panel.
func NewInfo(gui *Gui, name string, x, y, w, h int) *Info {
	return &Info{
		Gui:      gui,
		name:     name,
		Position: Position{x, y, w, h},
		Docker:   NewDockerInfo(gui),
		Host:     NewHostInfo(),
		Docui:    NewDocuiInfo(),
	}
}

// NewDocuiInfo create new docui info
func NewDocuiInfo() *Docui {
	return &Docui{
		Name:    "docui",
		Version: "1.0.2",
	}
}

// NewHostInfo create host info
func NewHostInfo() *HostInfo {
	return &HostInfo{
		OSType:       runtime.GOOS,
		Architecture: runtime.GOARCH,
	}
}

// NewDockerInfo create dockeri nfo
func NewDockerInfo(gui *Gui) *DockerInfo {
	info, err := gui.Docker.Info()
	if err != nil {
		return nil
	}

	var apiVersion string
	if v, err := gui.Docker.Version(); err != nil {
		apiVersion = ""
	} else {
		apiVersion = v.Get("ApiVersion")
	}

	return &DockerInfo{
		HostName:      info.Name,
		ServerVersion: info.ServerVersion,
		APIVersion:    apiVersion,
		KernelVersion: info.KernelVersion,
		OSType:        info.OSType,
		Architecture:  info.Architecture,
		Endpoint:      gui.Docker.Endpoint(),
		Containers:    info.Containers,
		Images:        info.Images,
		MemTotal:      fmt.Sprintf("%dMB", info.MemTotal/1024/1024),
	}
}
