package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
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
	client      *client.Client
	spec        func() (swarm.ServiceSpec, error)
	serviceName string
	imageName   string
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
	imageName := r.imageName

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

	if services, err := r.client.ServiceList(ctx, types.ServiceListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			"name",
			r.serviceName,
		}),
	}); err != nil {
		return fmt.Errorf("could not list swarm services: %s", err)
	} else {
		for _, service := range services {
			if err := r.client.ServiceRemove(ctx, service.ID); err != nil {
				fmt.Println("failed to remove service:", err)
			} else {
				fmt.Println("successfully removed service:", service.ID)
			}
		}
	}

	spec, err := r.spec()
	if err != nil {
		return err
	}

	if resp, err := r.client.ServiceCreate(ctx, spec, types.ServiceCreateOptions{
		QueryRegistry:       true,
		EncodedRegistryAuth: registryAuth,
	}); err != nil {
		return err
	} else {
		fmt.Println("Starting Docker Swarm service:", resp.ID)
		return nil
	}
}

func (d *dockerSwarm) RemoveContainer(ctx context.Context, containerName string) error {
	return d.client.ServiceRemove(ctx, getServiceId(containerName))
}

func getServiceId(input string) string {
	return "stack-" + input
}
