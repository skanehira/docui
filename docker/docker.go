package docker

import (
	"os"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

const (
	endpoint = "unix:///var/run/docker.sock"
)

type Docker struct {
	*docker.Client
}

func NewDocker() *Docker {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	return &Docker{client}
}

func (d *Docker) Images() []docker.APIImages {
	imgs, err := d.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		panic(err)
	}

	return imgs
}

func (d *Docker) Containers() []docker.APIContainers {
	cns, err := d.ListContainers(docker.ListContainersOptions{All: true})

	if err != nil {
		panic(err)
	}
	return cns
}

func (d *Docker) InspectContainer(name string) *docker.Container {
	con, err := d.Client.InspectContainer(name)
	if err != nil {
		panic(err)
	}

	return con
}

func (d *Docker) CreateContainerWithOptions(options docker.CreateContainerOptions) error {
	_, err := d.CreateContainer(options)
	if err != nil {
		return err
	}

	return nil
}

func (d *Docker) NewContainerOptions(config map[string]string) docker.CreateContainerOptions {

	options := docker.CreateContainerOptions{
		Config:     new(docker.Config),
		HostConfig: new(docker.HostConfig),
	}

	if name := config["Name"]; name != "" {
		options.Name = name
	}

	if image := config["Image"]; image != "" {
		options.Config.Image = image
	}

	if port := config["Port"]; port != "" {
		if hostPort := config["HostPort"]; hostPort != "" {
			options.HostConfig.PortBindings = map[docker.Port][]docker.PortBinding{
				docker.Port(port + "/tcp"): []docker.PortBinding{
					docker.PortBinding{
						HostIP:   "0.0.0.0",
						HostPort: hostPort,
					},
				},
			}
		}
	}

	if cmd := config["Cmd"]; cmd != "" {
		cmds := strings.Split(cmd, ",")
		for _, c := range cmds {
			options.Config.Cmd = append(options.Config.Cmd, c)
		}
	}

	if env := config["Env"]; env != "" {
		envs := strings.Split(env, ",")
		for _, v := range envs {
			options.Config.Env = append(options.Config.Env, v)
		}
	}

	if hostVolume := config["HostVolume"]; hostVolume != "" {
		if volume := config["Volume"]; volume != "" {
			options.HostConfig.Mounts = []docker.HostMount{
				docker.HostMount{
					Target: volume,
					Source: hostVolume,
					Type:   "bind",
				},
			}
		}
	}

	options.Config.Tty = true
	options.Config.AttachStdin = true
	options.Config.AttachStdout = true
	options.Config.AttachStderr = true
	options.Config.OpenStdin = true

	return options
}

func (d *Docker) CommitContainerWithOptions(options docker.CommitContainerOptions) error {
	if _, err := d.CommitContainer(options); err != nil {
		return err
	}
	return nil
}

func (d *Docker) RemoveContainerWithOptions(options docker.RemoveContainerOptions) error {
	if err := d.RemoveContainer(options); err != nil {
		return err
	}

	return nil
}

func (d *Docker) StartContainerWithID(id string) error {
	if err := d.StartContainer(id, nil); err != nil {
		return err
	}

	return nil
}

func (d *Docker) StopContainerWithID(id string) error {
	if err := d.StopContainer(id, 30); err != nil {
		return err
	}

	return nil
}

func (d *Docker) PullImageWithOptions(options docker.PullImageOptions) error {
	if err := d.PullImage(options, docker.AuthConfiguration{}); err != nil {
		return err
	}
	return nil
}

func (d *Docker) RemoveImageWithName(name string) error {
	if err := d.RemoveImage(name); err != nil {
		return err
	}

	return nil
}

func (d *Docker) SaveImageWithOptions(options docker.ExportImageOptions) error {
	if err := d.ExportImage(options); err != nil {
		return err
	}

	return nil
}

func (d *Docker) LoadImageWithPath(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	options := docker.LoadImageOptions{
		InputStream: file,
	}

	if err := d.LoadImage(options); err != nil {
		return err
	}

	return nil
}

func (d *Docker) ImportImageWithOptions(options docker.ImportImageOptions) error {

	if err := d.ImportImage(options); err != nil {
		return err
	}

	return nil
}

func (d *Docker) ExportContainerWithOptions(options docker.ExportContainerOptions) error {
	if err := d.ExportContainer(options); err != nil {
		return err
	}

	return nil
}

func (d *Docker) SearchImageWithName(name string) ([]docker.APIImageSearch, error) {
	images, err := d.Client.SearchImages(name)
	if err != nil {
		return images, err
	}

	return images, nil
}
