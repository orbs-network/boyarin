package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/runconfig"
	"io"
	"os"
)

type dockerRunner struct {
	client *client.Client

	containerName string
	config        map[string]interface{}
}

func NewDockerAPI() (Orchestrator, error) {
	client, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))

	if err != nil {
		return nil, err
	}

	return &dockerAPI{client: client}, nil
}

func (d *dockerAPI) PullImage(ctx context.Context, imageName string) error {
	out, err := d.client.ImagePull(ctx, imageName, types.ImagePullOptions{})

	if err != nil {
		return err
	}
	io.Copy(os.Stdout, out)

	return nil
}

func (r *dockerRunner) Run(ctx context.Context) error {
	jsonConfig, _ := json.Marshal(r.config)
	fmt.Println(string(jsonConfig))

	decoder := runconfig.ContainerDecoder{}
	config, hostConfig, networkConfig, err := decoder.DecodeConfig(bytes.NewReader(jsonConfig))
	if err != nil {
		return fmt.Errorf("could not parse Docker config: %s", err)
	}

	resp, err := r.client.ContainerCreate(ctx, config, hostConfig, networkConfig, r.containerName)
	if err != nil {
		return fmt.Errorf("could not create container: %s", err)
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
