package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/runconfig"
	"path/filepath"
)

type dockerRunner struct {
	client *client.Client

	containerName string
	config        map[string]interface{}
}

func NewDockerAPI(root string) (Orchestrator, error) {
	client, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))

	if err != nil {
		return nil, err
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	return &dockerAPI{
		client: client,
		root:   absoluteRoot,
	}, nil
}

func (d *dockerAPI) PullImage(ctx context.Context, imageName string) error {
	return pullImage(ctx, d.client, imageName)
}

func (r *dockerRunner) Run(ctx context.Context) error {
	jsonConfig, _ := json.Marshal(r.config)
	fmt.Println(string(jsonConfig))

	decoder := runconfig.ContainerDecoder{}
	config, hostConfig, networkConfig, err := decoder.DecodeConfig(bytes.NewReader(jsonConfig))
	if err != nil {
		return fmt.Errorf("could not parse Docker config: %s", err)
	}

	containers, err := r.client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(filters.KeyValuePair{
			"name",
			r.containerName,
		}),
	})

	if err != nil {
		return fmt.Errorf("could not get a list of containers: %s", err)
	}

	if len(containers) > 0 {
		for _, container := range containers {
			if err := r.client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
				Force: true,
			}); err != nil {
				fmt.Println("Failed to remove container:", err)
			} else {
				fmt.Println("Removed container:", container.ID)
			}
		}
	}

	resp, err := r.client.ContainerCreate(ctx, config, hostConfig, networkConfig, r.containerName)
	if err != nil {
		return fmt.Errorf("could not create contaiconner: %s", err)
	}

	if err := r.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("could not start container: %s", err)
	}

	fmt.Println("Started container:", resp.ID)
	return nil
}

func (d *dockerAPI) RemoveContainer(ctx context.Context, containerName string) error {
	return d.client.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: false,
	})
}
