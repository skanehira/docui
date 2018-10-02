package panel

import (
	"runtime"

	"github.com/jroimartin/gocui"
)

type Info struct {
	name   string
	Docker *DockerInfo
	Host   *HostInfo
}

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
	MemTotal      int64
}

type HostInfo struct {
	OSType       string
	Architecture string
}

func (info *Info) SetView(g *gocui.Gui) error {

	return nil
}

func (info *Info) Name() string {
	return info.name
}

func (info *Info) Refresh(g *gocui.Gui, v *gocui.View) error {
	return nil
}

func NewInfo(gui *Gui) *Info {
	return &Info{
		name:   InfoPanel,
		Docker: NewDockerInfo(gui),
		Host:   NewHostInfo(),
	}
}

func NewHostInfo() *HostInfo {
	return &HostInfo{
		OSType:       runtime.GOOS,
		Architecture: runtime.GOARCH,
	}
}

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
		MemTotal:      info.MemTotal,
	}
}
