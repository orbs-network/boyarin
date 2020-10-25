package adapter

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Only works for EFS volumes because they are shared
func (d *dockerSwarmOrchestrator) PurgeServiceData(ctx context.Context, containerName string) error {
	if d.options.StorageDriver != LOCAL_DRIVER && d.options.StorageMountType != "bind" {
		return fmt.Errorf("purging data for service %s not supported by the storage driver", containerName)
	}

	mounts, err := d.provisionServiceVolumes(ctx, containerName, nil)
	if err != nil {
		return err
	}

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

func (d *dockerSwarmOrchestrator) PurgeVchainData(ctx context.Context, nodeAddress string, containerName string) error {
	return nil
}
