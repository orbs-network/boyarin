package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"os"
)

type dockerSwarmOrchestrator struct {
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

func init() {
	os.Setenv("DOCKER_API_VERSION", DOCKER_API_VERSION)
}

func NewDockerSwarm(options OrchestratorOptions) (Orchestrator, error) {
	client, err := client.NewEnvClient()

	if err != nil {
		return nil, err
	}

	return &dockerSwarmOrchestrator{client: client, options: options}, nil
}

func (d *dockerSwarmOrchestrator) PullImage(ctx context.Context, imageName string) error {
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
		EncodedRegistryAuth: registryAuth,
	})

	return err
}

func (d *dockerSwarmOrchestrator) RemoveContainer(ctx context.Context, containerName string) error {
	if services, err := d.client.ServiceList(ctx, types.ServiceListOptions{
		Filters: FilterByName(containerName),
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

func GetServiceId(input string) string {
	return input + "-stack"
}

func (d *dockerSwarmOrchestrator) Close() error {
	return d.client.Close()
}
