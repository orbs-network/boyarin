package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"strconv"
	"strings"
)

const ORBS_BLOCKS_TARGET = "/usr/local/var/orbs"
const ORBS_LOGS_TARGET = "/opt/orbs/logs"

func getVolumeName(nodeAddress string, id uint32, postfix string) string {
	return fmt.Sprintf("%s-%d-%s", nodeAddress, id, postfix)
}

func (d *dockerSwarm) provisionVolumes(ctx context.Context, nodeAddress string, id uint32) (mounts []mount.Mount, err error) {
	if logsMount, err := d.provisionVolume(ctx, getVolumeName(nodeAddress, id, "logs"), ORBS_LOGS_TARGET, 2); err != nil {
		return mounts, err
	} else {
		mounts = append(mounts, logsMount)
	}

	if blocksMount, err := d.provisionVolume(ctx, getVolumeName(nodeAddress, id, "blocks"), ORBS_BLOCKS_TARGET, 8); err != nil {
		return mounts, err
	} else {
		mounts = append(mounts, blocksMount)
	}

	return mounts, nil
}

// FIXME propagate maxSize from the config
func (d *dockerSwarm) provisionVolume(ctx context.Context, volumeName string, target string, maxSizeInGb int) (mount.Mount, error) {
	driver := "local"
	if d.options.StorageDriver != "" {
		driver = d.options.StorageDriver
	}

	driverOptions := make(map[string]string)
	if len(d.options.StorageOptions) > 0 {
		driverOptions = d.options.StorageOptions
	}

	// Only enable size option for supported drivers
	if strings.HasPrefix(driver, "rexray/ebs") {
		driverOptions["size"] = strconv.Itoa(maxSizeInGb)
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
