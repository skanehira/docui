package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	volumetypes "github.com/docker/docker/api/types/volume"
)

// Volumes get volumes
func (d *Docker) Volumes() ([]*types.Volume, error) {
	res, err := d.VolumeList(context.TODO(), filters.Args{})
	if err != nil {
		return nil, err
	}

	return res.Volumes, nil
}

// InspectVolume inspect volume
func (d *Docker) InspectVolume(name string) (types.Volume, error) {
	volume, _, err := d.VolumeInspectWithRaw(context.TODO(), name)
	return volume, err
}

// RemoveVolume remove volume
func (d *Docker) RemoveVolume(name string) error {
	return d.VolumeRemove(context.TODO(), name, false)
}

// PruneVolumes remove unused volume
func (d *Docker) PruneVolumes() error {
	_, err := d.VolumesPrune(context.TODO(), filters.Args{})
	return err
}

// CreateVolume create volume
func (d *Docker) CreateVolume(opt volumetypes.VolumeCreateBody) error {
	_, err := d.VolumeCreate(context.TODO(), opt)
	return err
}

// NewCreateVolumeOptions generate options to create volume
func (d *Docker) NewCreateVolumeOptions(data map[string]string) volumetypes.VolumeCreateBody {
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

	return volumetypes.VolumeCreateBody{
		Name:       data["Name"],
		Driver:     data["Driver"],
		DriverOpts: driverOpts,
		Labels:     labels,
	}
}
