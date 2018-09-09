package main

import (
	docker "github.com/fsouza/go-dockerclient"
)

const (
	endpoint = "unix:///var/run/docker.sock"
)

type Docker struct {
	Endpoint string
	Client   *docker.Client
}

func NewDocker(endpoint string) *Docker {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	return &Docker{
		Endpoint: endpoint,
		Client:   client,
	}
}

func (d *Docker) Images() []docker.APIImages {
	imgs, err := d.Client.ListImages(docker.ListImagesOptions{All: true})
	if err != nil {
		panic(err)
	}

	return imgs
}

func (d *Docker) ImagesWithOptions(options docker.ListImagesOptions) []docker.APIImages {
	imgs, err := d.Client.ListImages(options)
	if err != nil {
		panic(err)
	}

	return imgs
}

func (d *Docker) InspectImage(name string) *docker.Image {
	img, err := d.Client.InspectImage(name)
	if err != nil {
		panic(err)
	}

	return img
}

func (d *Docker) Containers() []docker.APIContainers {
	cns, err := d.Client.ListContainers(docker.ListContainersOptions{All: true})

	if err != nil {
		panic(err)
	}
	return cns
}

func (d *Docker) ContainersWithOptions(options docker.ListContainersOptions) []docker.APIContainers {
	cns, err := d.Client.ListContainers(options)

	if err != nil {
		panic(err)
	}
	return cns
}

func (d *Docker) CreateContainerWithOptions(config map[string]string) error {
	_, err := d.Client.CreateContainer(NewContainerOptions(config))
	if err != nil {
		return err
	}

	return nil
}

func NewContainerOptions(config map[string]string) docker.CreateContainerOptions {

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
		options.Config.ExposedPorts = map[docker.Port]struct{}{
			docker.Port(port): struct{}{},
		}

		if hostPort := config["HostPort"]; hostPort != "" {
			options.HostConfig.PortBindings = map[docker.Port][]docker.PortBinding{
				docker.Port(port): []docker.PortBinding{
					docker.PortBinding{
						HostIP:   "0.0.0.0",
						HostPort: hostPort,
					},
				},
			}
		}
	}

	if env := config["Env"]; env != "" {
		options.Config.Env = []string{env}
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

	return options
}
