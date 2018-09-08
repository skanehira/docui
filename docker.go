package main

import docker "github.com/fsouza/go-dockerclient"

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
