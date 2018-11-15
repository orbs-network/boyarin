package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/runconfig"
	"io"
	"os"
	"time"
)

type dockerSwarm struct {
	client *client.Client
}

func NewDockerSwarm() (Orchestrator, error) {
	client, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))

	if err != nil {
		return nil, err
	}

	return &dockerSwarm{client: client}, nil
}

func (d *dockerSwarm) PullImage(ctx context.Context, imageName string) error {
	out, err := d.client.ImagePull(ctx, imageName, types.ImagePullOptions{})

	if err != nil {
		return err
	}
	io.Copy(os.Stdout, out)

	return nil
}

func (d *dockerSwarm) RunContainer(ctx context.Context, containerName string, dockerConfig interface{}) (string, error) {
	jsonConfig, _ := json.Marshal(dockerConfig)

	fmt.Println(string(jsonConfig))

	decoder := runconfig.ContainerDecoder{}
	config, _, _, err := decoder.DecodeConfig(bytes.NewReader(jsonConfig))
	if err != nil {
		return "", err
	}

	ureplicas := uint64(1)
	restartDelay := time.Duration(10 * time.Second)

	fmt.Println(config)

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Command: []string{"top"},
				Image:   config.Image,
				//Command: config.Cmd,
			},
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
		},
		Mode: swarm.ServiceMode{
			Replicated: &swarm.ReplicatedService{
				Replicas: &ureplicas,
			},
		},
	}
	spec.Name = getServiceId(containerName)

	resp, err := d.client.ServiceCreate(ctx, spec, types.ServiceCreateOptions{
		QueryRegistry: true,
	})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (d *dockerSwarm) RemoveContainer(ctx context.Context, containerName string) error {
	return d.client.ServiceRemove(ctx, getServiceId(containerName))
}

func (d *dockerSwarm) StoreConfiguration(ctx context.Context, containerName string, root string, config *AppConfig) error {
	panic("not implemented")
	return nil
}

func (d *dockerSwarm) GetContainerConfiguration(imageName string, containerName string, root string, httpPort int, gossipPort int) interface{} {
	panic("not implemented")
}

func getServiceId(input string) string {
	return "stack-" + input
}
