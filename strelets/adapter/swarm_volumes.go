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
	_, err := d.client.VolumeCreate(ctx, volume.VolumeCreateBody{
		Name:   volumeName,
		Driver: "local",
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
