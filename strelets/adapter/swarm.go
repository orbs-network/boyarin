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
	client  *client.Client
	options OrchestratorOptions
}

type dockerSwarmSecretsConfig struct {
	configSecretId  string
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

func NewDockerSwarm(options OrchestratorOptions) (Orchestrator, error) {
	client, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))

	if err != nil {
		return nil, err
	}

	return &dockerSwarm{client: client, options: options}, nil
}

func (d *dockerSwarm) PullImage(ctx context.Context, imageName string) error {
	return pullImage(ctx, d.client, imageName)
}

func (r *dockerSwarmRunner) Run(ctx context.Context) error {
	imageName := r.imageName

	var registryAuth string
	if username, password, err := getAuthForRepository(os.Getenv("HOME"), imageName); err != nil {
		// Ignore
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
				return fmt.Errorf("failed to remove service %s: %s", service.Spec.Name, err)
			}
		}
	}

	spec, err := r.spec()
	if err != nil {
		return err
	}

	_, err = r.client.ServiceCreate(ctx, spec, types.ServiceCreateOptions{
		QueryRegistry:       true,
		EncodedRegistryAuth: registryAuth,
	})

	return err
}

func (d *dockerSwarm) RemoveContainer(ctx context.Context, containerName string) error {
	return d.client.ServiceRemove(ctx, getServiceId(containerName))
}

func getServiceId(input string) string {
	return "stack-" + input
}

func (d *dockerSwarm) Close() error {
	return d.client.Close()
}
