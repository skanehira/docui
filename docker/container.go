package docker

import (
	"context"
	"io"
	"os"
	gosignal "os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/signal"
	"github.com/docker/go-connections/nat"
	"github.com/skanehira/docui/common"
	"github.com/skanehira/docui/docker/streams"
)

// CreateContainerOptions create container options
type CreateContainerOptions struct {
	Config        *container.Config
	HostConfig    *container.HostConfig
	NetworkConfig *network.NetworkingConfig
	Name          string
}

// Containers get containers
func (d *Docker) Containers(opt types.ContainerListOptions) ([]types.Container, error) {
	return d.ContainerList(context.TODO(), opt)
}

// InspectContainer inspect container
func (d *Docker) InspectContainer(name string) (types.ContainerJSON, error) {
	container, _, err := d.ContainerInspectWithRaw(context.TODO(), name, false)
	return container, err
}

// CreateContainer create container
func (d *Docker) CreateContainer(opt CreateContainerOptions) error {
	_, err := d.ContainerCreate(context.TODO(), opt.Config, opt.HostConfig, opt.NetworkConfig, opt.Name)
	return err
}

// NewContainerOptions generate container options to create container
func (d *Docker) NewContainerOptions(config map[string]string, isAttach bool) (CreateContainerOptions, error) {

	options := CreateContainerOptions{
		Config:     &container.Config{},
		HostConfig: &container.HostConfig{},
	}

	options.Config.Image = config["Image"]
	options.Name = config["Name"]

	image, _, err := d.ImageInspectWithRaw(context.TODO(), options.Config.Image)

	if user := config["User"]; user != "" {
		options.Config.User = user
	}

	if err != nil {
		return options, err
	}

	options.Config.Env = image.Config.Env

	port := config["Port"]
	hostPort := config["HostPort"]
	ip := config["HostIP"]

	if ip == "" {
		ip = "0.0.0.0"
	}

	if port != "" && hostPort != "" {
		options.HostConfig.PortBindings = nat.PortMap{
			nat.Port(port + "/tcp"): {
				{
					HostIP:   ip,
					HostPort: hostPort,
				},
			},
		}
	}

	if cmd := config["Cmd"]; cmd != "" {
		cmds := strings.Split(cmd, ",")
		for _, c := range cmds {
			options.Config.Cmd = append(options.Config.Cmd, c)
		}
	}

	if env := config["Env"]; env != "" {
		for _, v := range strings.Split(env, ",") {
			v = common.GetOSenv(v)
			options.Config.Env = append(options.Config.Env, v)
		}
	}

	hostVolume := config["HostVolume"]
	volume := config["Volume"]
	if hostVolume != "" && volume != "" {
		options.HostConfig.Mounts = []mount.Mount{
			{
				Target: volume,
				Source: hostVolume,
				Type:   mount.Type(config["VolumeType"]),
			},
		}
	}

	options.Config.AttachStdout = true
	options.Config.AttachStderr = true

	if isAttach {
		options.Config.Tty = true
		options.Config.AttachStdin = true
		options.Config.OpenStdin = true
	}

	return options, nil
}

// CommitContainer commit container
func (d *Docker) CommitContainer(name string, opt types.ContainerCommitOptions) error {
	_, err := d.ContainerCommit(context.TODO(), name, opt)
	return err
}

// RemoveContainer remove container
func (d *Docker) RemoveContainer(name string) error {
	return d.ContainerRemove(context.TODO(), name, types.ContainerRemoveOptions{})
}

// RenameContainer rename container
func (d *Docker) RenameContainer(id, newName string) error {
	return d.ContainerRename(context.TODO(), id, newName)
}

// StartContainer start container with id
func (d *Docker) StartContainer(id string) error {
	return d.ContainerStart(context.TODO(), id, types.ContainerStartOptions{})
}

// StopContainer stop container with id
func (d *Docker) StopContainer(id string) error {
	return d.ContainerStop(context.TODO(), id, nil)
}

// ExportContainer export container
func (d *Docker) ExportContainer(name, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	out, err := d.ContainerExport(context.TODO(), name)

	if err != nil {
		return err
	}

	if _, err = io.Copy(file, out); err != nil {
		return err
	}

	return nil
}

// CreateExec container exec create
func (d *Docker) CreateExec(container, cmd string) (types.IDResponse, error) {
	return d.ContainerExecCreate(context.TODO(), container, types.ExecConfig{
		Tty:          true,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          []string{cmd},
	})
}

// AttachExecContainer attach container
func (d *Docker) AttachExecContainer(id, cmd string) error {
	exec, err := d.CreateExec(id, cmd)

	if err != nil {
		common.Logger.Error(err)
		return err
	}

	ctx := context.TODO()

	resp, err := d.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		common.Logger.Error(err)
		return err
	}

	defer resp.Close()

	errCh := make(chan error)
	std := streams.NewStd()

	go func() {
		defer close(errCh)
		errCh <- func() error {
			streamer := hijackedIOStreamer{
				streams:      std,
				inputStream:  std.In(),
				outputStream: std.Out(),
				errorStream:  std.Err(),
				resp:         resp,
				tty:          true,
			}

			return streamer.stream(ctx)
		}()
	}()

	if std.In().IsTerminal() {
		if err := monitorTtySize(ctx, d, std, exec.ID, true); err != nil {
			// output error log
		}
	}
	if err := <-errCh; err != nil {
		common.Logger.Error(err)
		return err
	}
	return nil
}

func resizeTtyTo(ctx context.Context, docker *Docker, id string, height, width uint, isExec bool) {
	if height == 0 && width == 0 {
		return
	}

	options := types.ResizeOptions{
		Height: height,
		Width:  width,
	}

	var err error
	if isExec {
		err = docker.ContainerExecResize(ctx, id, options)
	} else {
		err = docker.ContainerResize(ctx, id, options)
	}

	if err != nil {
		// output error log
	}
}

func monitorTtySize(ctx context.Context, docker *Docker, std *streams.Std, id string, isExec bool) error {
	resizeTty := func() {
		height, width := std.Out().GetTtySize()
		resizeTtyTo(ctx, docker, id, height, width, isExec)
	}

	resizeTty()

	if runtime.GOOS == "windows" {
		go func() {
			prevH, prevW := std.Out().GetTtySize()
			for {
				time.Sleep(time.Millisecond * 250)
				h, w := std.Out().GetTtySize()

				if prevW != w || prevH != h {
					resizeTty()
				}
				prevH = h
				prevW = w
			}
		}()
	} else {
		sigchan := make(chan os.Signal, 1)
		gosignal.Notify(sigchan, signal.SIGWINCH)
		go func() {
			for range sigchan {
				resizeTty()
			}
		}()
	}
	return nil
}
