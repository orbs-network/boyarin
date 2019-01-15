package adapter

import (
	"context"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
)

func (d *dockerSwarm) provisionVolumes(ctx context.Context, containerName string) (mounts []mount.Mount, err error) {
	if logsMount, err := d.provisionVolume(ctx, containerName+"-logs", "/opt/orbs/logs"); err != nil {
		return mounts, err
	} else {
		mounts = append(mounts, logsMount)
	}

	if blocksMount, err := d.provisionVolume(ctx, containerName+"-blocks", "/usr/local/var/orbs"); err != nil {
		return mounts, err
	} else {
		mounts = append(mounts, blocksMount)
	}

	return mounts, nil
}

func (d *dockerSwarm) provisionVolume(ctx context.Context, volumeName string, target string) (mount.Mount, error) {
	driver := "local"
	if d.options.StorageDriver() != "" {
		driver = d.options.StorageDriver()
	}

	driverOptions := make(map[string]string)
	if len(d.options.StorageOptions()) > 0 {
		driverOptions = d.options.StorageOptions()
	}

	_, err := d.client.VolumeCreate(ctx, volume.VolumeCreateBody{
		Name:       volumeName,
		Driver:     driver,
		DriverOpts: driverOptions,
	})

	if err != nil {
		return mount.Mount{}, err
	}

	return mount.Mount{
		Source:   volumeName,
		Type:     "volume",
		Target:   target,
		ReadOnly: false,
	}, nil
}
