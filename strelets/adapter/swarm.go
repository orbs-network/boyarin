package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
	"os"
)

type dockerSwarmOrchestrator struct {
	client  *client.Client
	options OrchestratorOptions
	logger  log.Logger
}

type dockerSwarmSecretsConfig struct {
	configSecretId  string
	networkSecretId string
	keysSecretId    string
}

type dockerSwarmNginxSecretsConfig struct {
	nginxConfId      string
	vchainConfId     string
	sslCertificateId string
	sslPrivateKeyId  string
}

func NewDockerSwarm(options OrchestratorOptions, logger log.Logger) (Orchestrator, error) {
	client, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))

	if err != nil {
		return nil, err
	}

	return &dockerSwarmOrchestrator{client: client, options: options, logger: logger}, nil
}

func (d *dockerSwarmOrchestrator) PullImage(ctx context.Context, imageName string) error {
	return pullImage(ctx, d.client, imageName)
}

func (d *dockerSwarmOrchestrator) create(ctx context.Context, spec swarm.ServiceSpec, imageName string) error {
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
	_, err := d.client.ServiceCreate(ctx, spec, types.ServiceCreateOptions{
		QueryRegistry:       true,
		EncodedRegistryAuth: registryAuth,
	})

	return errors.Wrap(err, "failed creating service")
}

func (d *dockerSwarmOrchestrator) RemoveService(ctx context.Context, serviceName string) error {
	services, err := d.client.ServiceList(ctx, types.ServiceListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "name", Value: serviceName}),
	})
	if err != nil {
		return fmt.Errorf("could not list swarm services: %s \n %v", serviceName, err)
	}
	if len(services) == 0 {
		d.logger.Info(fmt.Sprintf("no service found for removal: %s", serviceName))
		return nil
	}
	for _, service := range services {
		if err := d.client.ServiceRemove(ctx, service.ID); err != nil {
			return fmt.Errorf("failed to remove service %s with id %s", serviceName, service.ID)
		} else {
			d.logger.Info(fmt.Sprintf("successfully removed service %s with id %s", serviceName, service.ID))
		}
	}
	return nil
}

func (d *dockerSwarmOrchestrator) Close() error {
	return d.client.Close()
}
