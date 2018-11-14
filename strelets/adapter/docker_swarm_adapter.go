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

func NewDockerSwarm() (DockerAPI, error) {
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

func (d *dockerSwarm) RunContainer(ctx context.Context, containerName string, dockerConfig map[string]interface{}) (string, error) {
	jsonConfig, _ := json.Marshal(dockerConfig)

	fmt.Println(string(jsonConfig))

	decoder := runconfig.ContainerDecoder{}
	config, _, _, err := decoder.DecodeConfig(bytes.NewReader(jsonConfig))
	if err != nil {
		return "", err
	}

	fmt.Println(containerName, config.Image)

	fmt.Println(d.client.ServiceRemove(ctx, "top"))

	ureplicas := uint64(1)
	restartDelay := time.Duration(100 * time.Millisecond)

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   "busybox:latest",
				Command: []string{"/bin/top"},
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
	spec.Name = "stack-" + containerName

	resp, err := d.client.ServiceCreate(ctx, spec, types.ServiceCreateOptions{
		QueryRegistry: true,
	})
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (d *dockerSwarm) RemoveContainer(ctx context.Context, containerName string) error {
	//return d.client.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
	//	Force:         true,
	//	RemoveLinks:   false,
	//	RemoveVolumes: false,
	//})

	return nil
}
