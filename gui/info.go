package gui

import (
	"context"
	"fmt"
	"runtime"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/skanehira/docui/docker"
)

type info struct {
	*tview.TextView
	Docker *dockerInfo
	Host   *hostInfo
	Docui  *docui
}

type docui struct {
	Name    string
	Version string
}

type dockerInfo struct {
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

type hostInfo struct {
	OSType       string
	Architecture string
}

func newDocuiInfo() *docui {
	return &docui{
		Name:    "docui",
		Version: "2.0.0",
	}
}

func newHostInfo() *hostInfo {
	return &hostInfo{
		OSType:       runtime.GOOS,
		Architecture: runtime.GOARCH,
	}
}

func newDockerInfo() *dockerInfo {
	info, err := docker.Client.Info(context.TODO())
	if err != nil {
		return nil
	}

	var apiVersion string
	if v, err := docker.Client.ServerVersion(context.TODO()); err != nil {
		apiVersion = ""
	} else {
		apiVersion = v.APIVersion
	}

	return &dockerInfo{
		HostName:      info.Name,
		ServerVersion: info.ServerVersion,
		APIVersion:    apiVersion,
		KernelVersion: info.KernelVersion,
		OSType:        info.OSType,
		Architecture:  info.Architecture,
		Endpoint:      docker.Client.DaemonHost(),
		Containers:    info.Containers,
		Images:        info.Images,
		MemTotal:      fmt.Sprintf("%dMB", info.MemTotal/1024/1024),
	}
}

func newInfo() *info {
	i := &info{
		TextView: tview.NewTextView(),
		Docker:   newDockerInfo(),
		Host:     newHostInfo(),
		Docui:    newDocuiInfo(),
	}

	i.display()

	return i
}

func (i *info) display() {
	dockerAPI := fmt.Sprintf("api version:%s", i.Docker.APIVersion)
	dockerVersion := fmt.Sprintf("server version:%s", i.Docker.ServerVersion)
	dockerEndpoint := fmt.Sprintf("endpoint:%s", i.Docker.Endpoint)
	docuiVersion := fmt.Sprintf("version:%s", i.Docui.Version)

	i.SetTextColor(tcell.ColorYellow)
	i.SetText(fmt.Sprintf(" docker	| %s %s %s\n docui	 | %s", dockerAPI, dockerVersion, dockerEndpoint, docuiVersion))
}
