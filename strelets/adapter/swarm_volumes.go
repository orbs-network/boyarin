package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"os"
	"strconv"
	"strings"
)

const ORBS_BLOCKS_TARGET = "/usr/local/var/orbs"
const ORBS_LOGS_TARGET = "/opt/orbs/logs"

const REXRAY_EBS_DRIVER = "rexray/ebs"
const LOCAL_DRIVER = "local"

func getVolumeName(nodeAddress string, id uint32, postfix string) string {
	return fmt.Sprintf("%s-%d-%s", nodeAddress, id, postfix)
}

func (d *dockerSwarmOrchestrator) provisionVolumes(ctx context.Context, nodeAddress string, id uint32, blocksVolumeSize int, logsVolumeSize int) (mounts []mount.Mount, err error) {
	if logsMount, err := d.provisionVolume(ctx, getVolumeName(nodeAddress, id, "logs"), ORBS_LOGS_TARGET, logsVolumeSize); err != nil {
		return mounts, err
	} else {
		mounts = append(mounts, logsMount)
	}

	if blocksMount, err := d.provisionVolume(ctx, getVolumeName(nodeAddress, id, "blocks"), ORBS_BLOCKS_TARGET, blocksVolumeSize); err != nil {
		return mounts, err
	} else {
		mounts = append(mounts, blocksMount)
	}

	return mounts, nil
}

func (d *dockerSwarmOrchestrator) provisionVolume(ctx context.Context, volumeName string, target string, maxSizeInGb int) (mount.Mount, error) {
	source, driverName, driverOptions := getVolumeDriverOptions(volumeName, d.options, maxSizeInGb)

	_, err := d.client.VolumeCreate(ctx, volume.VolumeCreateBody{
		Name:       volumeName,
		Driver:     driverName,
		DriverOpts: driverOptions,
	})

	if err != nil {
		return mount.Mount{}, err
	}

	return mount.Mount{
		Source:   source,
		Type:     d.options.MountType(),
		Target:   target,
		ReadOnly: false,
		VolumeOptions: getVolumeOptions(d.options, driverName, driverOptions),
	}, nil
}

func getVolumeOptions(orchestratorOptions OrchestratorOptions, driverName string, driverOptions map[string]string) *mount.VolumeOptions {
	switch orchestratorOptions.MountType() {
	case mount.TypeVolume:
		return &mount.VolumeOptions{
			DriverConfig: &mount.Driver{
				Name:    driverName,
				Options: driverOptions,
			},
		}
	}

	return nil
}

func getVolumeDriverOptions(volumeName string, orchestratorOptions OrchestratorOptions, maxSizeInGb int) (string, string, map[string]string) {
	driver := LOCAL_DRIVER

	if orchestratorOptions.StorageDriver != "" {
		driver = orchestratorOptions.StorageDriver
	}

	driverOptions := make(map[string]string)
	for k, v := range orchestratorOptions.StorageOptions {
		driverOptions[k] = v
	}

	// Only enable size option for supported drivers
	switch driver {
	case REXRAY_EBS_DRIVER:
		driverOptions["size"] = strconv.Itoa(maxSizeInGb)
	case LOCAL_DRIVER:
		switch orchestratorOptions.MountType() {
		case mount.TypeVolume:
			if fsType, ok := driverOptions["type"]; ok && fsType == "nfs" {
				// append volumeName to the common shared volume storage directory
				dir := driverOptions["device"] + "/" + volumeName
				driverOptions["device"] = dir
				// Warning: we assume that the volume directory exists on this machine, or its parent is mounted
				if strings.HasPrefix(dir, ":") {
					dir = dir[1:]
				}
				os.MkdirAll(dir, 0755)
			}
		case mount.TypeBind:
			volumeName = "/var/efs/" + volumeName
			os.MkdirAll(volumeName, 0755)
		}
	}

	return volumeName, driver, driverOptions
}
