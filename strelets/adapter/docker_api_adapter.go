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

const DOCKER_API_VERSION = "1.38"

type dockerAPI struct {
	client *client.Client
}

func NewDockerAPI() (DockerAPI, error) {
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

func (d *dockerAPI) RunContainer(ctx context.Context, imageName string, containerName string, dockerConfig map[string]interface{}) (string, error) {
	jsonConfig, _ := json.Marshal(dockerConfig)

	fmt.Println(string(jsonConfig))

	decoder := runconfig.ContainerDecoder{}
	config, hostConfig, networkConfig, err := decoder.DecodeConfig(bytes.NewReader(jsonConfig))
	if err != nil {
		return "", err
	}

	resp, err := d.client.ContainerCreate(ctx, config, hostConfig, networkConfig, containerName)
	if err != nil {
		return "", err
	}

	if err := d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (d *dockerAPI) RemoveContainer(ctx context.Context, containerName string) error {
	return d.client.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: false,
	})
}
