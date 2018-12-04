package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"os"
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

type dockerSwarmNginxSecretsConfig struct {
	nginxConfId  string
	vchainConfId string
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
	imageName := r.spec.TaskTemplate.ContainerSpec.Image

	var registryAuth string
	if username, password, err := getAuthForRepository(os.Getenv("HOME"), imageName); err != nil {
		fmt.Println(err)
	} else {
		registryAuth = encodeAuthConfig(&types.AuthConfig{
			Username:      username,
			Password:      password,
			ServerAddress: getRepoName(imageName),
		})
	}

	if resp, err := r.client.ServiceCreate(ctx, r.spec, types.ServiceCreateOptions{
		QueryRegistry:       true,
		EncodedRegistryAuth: registryAuth,
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
