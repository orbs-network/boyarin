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
	nginxConfId      string
	vchainConfId     string
	sslCertificateId string
	sslPrivateKeyId  string
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
	if services, err := d.client.ServiceList(ctx, types.ServiceListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{"name", containerName}),
	}); err != nil {
		return fmt.Errorf("could not list swarm services: %s", err)
	} else {
		for _, service := range services {
			if err := d.client.ServiceRemove(ctx, service.ID); err != nil {
				return fmt.Errorf("failed to remove service %s with id %s", containerName, service.ID)
			}
		}
	}

	return nil
}

func getServiceId(input string) string {
	return input + "-stack"
}

func (d *dockerSwarm) Close() error {
	return d.client.Close()
}
