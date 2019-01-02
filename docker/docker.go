package docker

import (
	"errors"
	"fmt"
	"os"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/skanehira/docui/common"
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

func (d *Docker) Images(options docker.ListImagesOptions) []docker.APIImages {
	imgs, err := d.ListImages(options)
	if err != nil {
		return []docker.APIImages{}
	}

	return imgs
}

func (d *Docker) Containers() []docker.APIContainers {
	cns, err := d.ListContainers(docker.ListContainersOptions{All: true})

	if err != nil {
		return []docker.APIContainers{}
	}
	return cns
}

func (d *Docker) Networks() []docker.Network {
	net, err := d.ListNetworks()
	if err != nil {
		return []docker.Network{}
	}

	return net
}

func (d *Docker) CreateContainerWithOptions(options docker.CreateContainerOptions) error {
	_, err := d.CreateContainer(options)
	if err != nil {
		return err
	}

	return nil
}

func (d *Docker) NewContainerOptions(config map[string]string, isAttach bool) (docker.CreateContainerOptions, error) {

	options := docker.CreateContainerOptions{
		Config:     new(docker.Config),
		HostConfig: new(docker.HostConfig),
	}

	if image := config["Image"]; image != "" {
		options.Config.Image = image
	} else {
		return options, fmt.Errorf("no specified image")
	}

	if name := config["Name"]; name != "" {
		options.Name = name
	}

	image, err := d.InspectImage(options.Config.Image)

	if user := config["User"]; user != "" {
		options.Config.User = user
	}

	if err != nil {
		return options, err
	}

	options.Config.Env = image.Config.Env

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
		for _, v := range strings.Split(env, ",") {
			v = common.GetOSenv(v)
			options.Config.Env = append(options.Config.Env, v)
		}
	}

	hostVolume := config["HostVolume"]
	volume := config["Volume"]

	if hostVolume != "" && volume == "" {
		return options, fmt.Errorf("no specified Volume")
	}
	if hostVolume == "" && volume != "" {
		return options, fmt.Errorf("no specified HostVoluem")
	}

	if hostVolume != "" && volume != "" {
		options.HostConfig.Mounts = []docker.HostMount{
			docker.HostMount{
				Target: volume,
				Source: hostVolume,
				Type:   config["VolumeType"],
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

func (d *Docker) RenameContainerWithOptions(options docker.RenameContainerOptions) error {
	if err := d.RenameContainer(options); err != nil {
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

func (d *Docker) RemoveDanglingImages() error {
	options := docker.ListImagesOptions{
		Filters: map[string][]string{
			"dangling": []string{
				"true", "1",
			},
		},
	}

	images := d.Images(options)
	errids := []string{}

	for _, image := range images {
		if err := d.RemoveImageWithName(image.ID); err != nil {
			errids = append(errids, image.ID[7:19])
		}
	}

	if len(errids) > 1 {
		return errors.New(fmt.Sprintf("can not remove ids\n%s", errids))
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

func (d *Docker) Volumes() []docker.Volume {
	volumes, err := d.Client.ListVolumes(docker.ListVolumesOptions{})

	if err != nil {
		return volumes
	}

	return volumes
}

func (d *Docker) RemoveVolumeWithName(name string) error {
	return d.RemoveVolume(name)
}

func (d *Docker) PruneVolumes() error {
	_, err := d.Client.PruneVolumes(docker.PruneVolumesOptions{})

	if err != nil {
		return err
	}

	return nil
}

func (d *Docker) CreateVolumeWithOptions(options docker.CreateVolumeOptions) error {
	_, err := d.Client.CreateVolume(options)
	return err
}

func (d *Docker) NewCreateVolumeOptions(data map[string]string) docker.CreateVolumeOptions {
	driverOpts := make(map[string]string)
	labels := make(map[string]string)

	for _, label := range strings.Split(data["Labels"], " ") {
		kv := strings.SplitN(label, "=", 2)

		if len(kv) > 1 && kv[1] != "" {
			labels[kv[0]] = kv[1]
		}
	}

	for _, opt := range strings.Split(data["Options"], " ") {
		kv := strings.SplitN(opt, "=", 2)

		if len(kv) > 1 && kv[1] != "" {
			driverOpts[kv[0]] = kv[1]
		}
	}

	options := docker.CreateVolumeOptions{
		Name:       data["Name"],
		Driver:     data["Driver"],
		DriverOpts: driverOpts,
		Labels:     labels,
	}

	return options
}

func (d *Docker) DiskUsage() *docker.DiskUsage {
	usage, err := d.Client.DiskUsage(docker.DiskUsageOptions{})

	if err != nil {
		return nil
	}

	return usage
}
