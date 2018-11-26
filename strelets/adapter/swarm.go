package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

type dockerSwarm struct {
	client *client.Client
}

type dockerSwarmSecretsConfig struct {
	networkSecretId string
	keysSecretId    string
}

type dockerSwarmRunner struct {
	client *client.Client
	spec   swarm.ServiceSpec
}

func NewDockerSwarm() (Orchestrator, error) {
	client, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))

	if err != nil {
		return nil, err
	}

	return &dockerSwarm{client: client}, nil
}

func (d *dockerSwarm) PullImage(ctx context.Context, imageName string) error {
	return pullImage(ctx, d.client, imageName)
}

func (r *dockerSwarmRunner) Run(ctx context.Context) error {
	if resp, err := r.client.ServiceCreate(ctx, r.spec, types.ServiceCreateOptions{
		QueryRegistry: true,
	}); err != nil {
		return err
	} else {
		fmt.Println("Starting Docker Swarm stack:", resp.ID)
		return nil
	}
}

func (d *dockerSwarm) RemoveContainer(ctx context.Context, containerName string) error {
	return d.client.ServiceRemove(ctx, getServiceId(containerName))
}

func getServiceId(input string) string {
	return "stack-" + input
}

func (d *dockerSwarm) PrepareReverseProxy(ctx context.Context, config string) (Runner, error) {
	panic("not implemented")
	return nil, nil
}
