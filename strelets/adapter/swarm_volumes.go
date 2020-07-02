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
const ORBS_STATUS_TARGET = "/opt/orbs/status"
const ORBS_CACHE_TARGET = "/opt/orbs/cache"

const REXRAY_EBS_DRIVER = "rexray/ebs"
const LOCAL_DRIVER = "local"

func getVchainVolumeName(nodeAddress string, vcId uint32, postfix string) string {
	return fmt.Sprintf("%s-%d-%s", nodeAddress, vcId, postfix)
}

func getServiceVolumeName(serviceName string, postfix string) string {
	return fmt.Sprintf("%s-%s", serviceName, postfix)
}

func (d *dockerSwarmOrchestrator) provisionVchainVolume(ctx context.Context, nodeAddress string, vcId uint32, blocksVolumeSize int) (mount.Mount, error) {
	return d.provisionVolume(ctx, getVchainVolumeName(nodeAddress, vcId, "blocks"), ORBS_BLOCKS_TARGET, blocksVolumeSize, d.options)
}

func (d *dockerSwarmOrchestrator) provisionLogsVolume(ctx context.Context, nodeAddress string, serviceName string, mountTarget string, logsVolumeSize int) (mount.Mount, error) {
	return d.provisionVolume(ctx, getServiceVolumeName(serviceName, "logs"), mountTarget, logsVolumeSize, OrchestratorOptions{})
}

func (d *dockerSwarmOrchestrator) provisionStatusVolume(ctx context.Context, nodeAddress string, serviceName string, mountTarget string) (mount.Mount, error) {
	return d.provisionVolume(ctx, getServiceVolumeName(serviceName, "status"), mountTarget, 0, d.options)
}

func (d *dockerSwarmOrchestrator) provisionCacheVolume(ctx context.Context, nodeAddress string, serviceName string) (mount.Mount, error) {
	return d.provisionVolume(ctx, getServiceVolumeName(serviceName, "cache"), ORBS_CACHE_TARGET, 0, d.options)
}

func (d *dockerSwarmOrchestrator) provisionVolume(ctx context.Context, volumeName string, target string, maxSizeInGb int, orchestratorOptions OrchestratorOptions) (mount.Mount, error) {
	source, driverName, driverOptions := getVolumeDriverOptions(volumeName, orchestratorOptions, maxSizeInGb)

	_, err := d.client.VolumeCreate(ctx, volume.VolumeCreateBody{
		Name:       volumeName,
		Driver:     driverName,
		DriverOpts: driverOptions,
	})

	if err != nil {
		return mount.Mount{}, err
	}

	return mount.Mount{
		Source:        source,
		Type:          orchestratorOptions.MountType(),
		Target:        target,
		ReadOnly:      false,
		VolumeOptions: getVolumeOptions(orchestratorOptions, driverName, driverOptions),
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

	fmt.Println("Allocating driver options for a volume")
	fmt.Println("orch options: ", orchestratorOptions)
	fmt.Println("volume name", volumeName)

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

	fmt.Println("about to return the following volume options", volumeName, driver, driverOptions)

	return volumeName, driver, driverOptions
}
