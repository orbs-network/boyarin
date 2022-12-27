package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/pkg/errors"
	"os"
	"strings"
)

const ORBS_BLOCKS_TARGET = "/usr/local/var/orbs"
const ORBS_LOGS_TARGET = "/opt/orbs/logs"
const ORBS_STATUS_TARGET = "/opt/orbs/status"
const ORBS_CACHE_TARGET = "/opt/orbs/cache"

const REXRAY_EBS_DRIVER = "rexray/ebs"
const LOCAL_DRIVER = "local"

const DEFAULT_EFS_PATH = "/var/efs/"

func getServiceVolumeName(serviceName string, postfix string) string {
	return fmt.Sprintf("%s-%s", serviceName, postfix)
}

func (d *dockerSwarmOrchestrator) provisionLogsVolume(ctx context.Context, serviceName string, mountTarget string) (mount.Mount, error) {
	return d.provisionVolume(ctx, getServiceVolumeName(serviceName, "logs"), mountTarget, d.options)
}

func (d *dockerSwarmOrchestrator) provisionStatusVolume(ctx context.Context, serviceName string, mountTarget string) (mount.Mount, error) {
	return d.provisionVolume(ctx, getServiceVolumeName(serviceName, "status"), mountTarget, d.options)
}

func (d *dockerSwarmOrchestrator) provisionCacheVolume(ctx context.Context, serviceName string) (mount.Mount, error) {
	return d.provisionVolume(ctx, getServiceVolumeName(serviceName, "cache"), ORBS_CACHE_TARGET, d.options)
}

func (d *dockerSwarmOrchestrator) provisionVolume(ctx context.Context, volumeName string, target string, orchestratorOptions *OrchestratorOptions) (mount.Mount, error) {
	if orchestratorOptions.StorageDriver == REXRAY_EBS_DRIVER {
		return mount.Mount{}, errors.Errorf("%s storage driver is no longer supported, please consult how to enable EFS instead", REXRAY_EBS_DRIVER)
	}

	driverName := LOCAL_DRIVER
	source, driverOptions := getVolumeDriverOptions(volumeName, orchestratorOptions)

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

func getVolumeOptions(orchestratorOptions *OrchestratorOptions, driverName string, driverOptions map[string]string) *mount.VolumeOptions {
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

func getVolumeDriverOptions(volumeName string, orchestratorOptions *OrchestratorOptions) (string, map[string]string) {
	driverOptions := make(map[string]string)
	for k, v := range orchestratorOptions.StorageOptions {
		driverOptions[k] = v
	}

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
		volumeName = DEFAULT_EFS_PATH + volumeName
		os.MkdirAll(volumeName, 0755)
	}

	return volumeName, driverOptions
}
