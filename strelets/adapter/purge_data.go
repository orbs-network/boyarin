package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/orbs-network/boyarin/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Only works for EFS volumes because they are shared
func (d *dockerSwarmOrchestrator) canPurgeData(containerName string) error {
	if d.options.StorageDriver != LOCAL_DRIVER && d.options.StorageMountType != "bind" {
		return fmt.Errorf("purging data for service %s not supported by the storage driver", containerName)
	}

	return nil
}

func (d *dockerSwarmOrchestrator) purgeMounts(mounts []mount.Mount) error {
	var errors []error
	for _, m := range mounts {
		volumeName := m.Source

		if fileInfos, err := ioutil.ReadDir(volumeName); err != nil {
			errors = append(errors, err)
		} else {
			for _, fileInfo := range fileInfos {
				if err := os.RemoveAll(filepath.Join(volumeName, fileInfo.Name())); err != nil {
					errors = append(errors, err)
				}
			}
		}
	}

	return utils.AggregateErrors(errors)
}

func (d *dockerSwarmOrchestrator) PurgeServiceData(ctx context.Context, containerName string) error {
	if err := d.canPurgeData(containerName); err != nil {
		return err
	}

	mounts, err := d.provisionServiceVolumes(ctx, containerName, nil)
	if err != nil {
		return err
	}

	return d.purgeMounts(mounts)
}

func (d *dockerSwarmOrchestrator) PurgeVirtualChainData(ctx context.Context, nodeAddress string, vcId uint32, containerName string) error {
	if err := d.canPurgeData(containerName); err != nil {
		return err
	}

	mounts, err := d.provisionServiceVolumes(ctx, containerName, nil)
	if err != nil {
		return err
	}

	if blocksMount, err := d.provisionVchainVolume(ctx, nodeAddress, vcId); err != nil {
		return fmt.Errorf("failed to access volumes: %s", err)
	} else {
		mounts = append(mounts, blocksMount)
	}

	return d.purgeMounts(mounts)
}
