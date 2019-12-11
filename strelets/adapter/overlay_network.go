package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
)

const SHARED_SIGNER_NETWORK = "signer-overlay"

func (d *dockerSwarmOrchestrator) GetOverlayNetwork(ctx context.Context, name string) (string, error) {
	networks, err := d.client.NetworkList(ctx, types.NetworkListOptions{
		Filters: FilterByName(name),
	})

	if err != nil {
		return "", fmt.Errorf("could not list networks: %s", err)
	}

	if len(networks) == 0 {
		response, err := d.client.NetworkCreate(ctx, name, types.NetworkCreate{
			Driver:         "overlay",
			Attachable:     true,
			CheckDuplicate: true,
		})

		if err != nil {
			return "", fmt.Errorf("could not create overlay network %s: %s", name, err)
		}

		return response.ID, nil
	}

	return networks[0].ID, nil
}
