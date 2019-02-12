package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/registry"
)

// Images get images from
func (d *Docker) Images(opt types.ImageListOptions) ([]types.ImageSummary, error) {
	return d.ImageList(context.TODO(), opt)
}

// InspectImage inspect image
func (d *Docker) InspectImage(name string) (types.ImageInspect, error) {
	img, _, err := d.ImageInspectWithRaw(context.TODO(), name)
	return img, err
}

// PullImage pull image
func (d *Docker) PullImage(name string) error {
	resp, err := d.ImagePull(context.TODO(), name, types.ImagePullOptions{})

	if err != nil {
		return err
	}

	// wait until pull is completed
	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// RemoveImage remove image
func (d *Docker) RemoveImage(name string) error {
	_, err := d.ImageRemove(context.TODO(), name, types.ImageRemoveOptions{})
	return err
}

// RemoveDanglingImages remove dangling images
func (d *Docker) RemoveDanglingImages() error {
	opt := types.ImageListOptions{
		Filters: filters.NewArgs(filters.Arg("dangling", "true")),
	}

	images, err := d.Images(opt)
	if err != nil {
		return err
	}

	errIDs := []string{}

	for _, image := range images {
		if err := d.RemoveImage(image.ID); err != nil {
			errIDs = append(errIDs, image.ID[7:19])
		}
	}

	if len(errIDs) > 1 {
		return fmt.Errorf("can not remove ids\n%s", errIDs)
	}

	return nil
}

// SaveImage save image to tar file
func (d *Docker) SaveImage(ids []string, path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	out, err := d.ImageSave(context.TODO(), ids)
	if err != nil {
		return err
	}

	if _, err = io.Copy(file, out); err != nil {
		return err
	}

	return nil
}

// LoadImage load image from tar file
func (d *Docker) LoadImage(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = d.ImageLoad(context.TODO(), file, true)
	return err
}

// ImportImage import image
func (d *Docker) ImportImage(name, tag, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	source := types.ImageImportSource{
		Source:     file,
		SourceName: "-",
	}

	opt := types.ImageImportOptions{
		Tag: tag,
	}

	_, err = d.ImageImport(context.TODO(), source, name, opt)

	return err
}

// SearchImage search images
func (d *Docker) SearchImage(name string) ([]registry.SearchResult, error) {
	// https://github.com/moby/moby/blob/8e610b2b55bfd1bfa9436ab110d311f5e8a74dcb/registry/service.go#L22
	// Limit default:25 min:1 max:100
	return d.ImageSearch(context.TODO(), name, types.ImageSearchOptions{Limit: 100})
}
