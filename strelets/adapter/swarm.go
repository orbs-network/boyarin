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
	storedSecrets, err := d.storeNginxConfiguration(ctx, config)
	if err != nil {
		return nil, err
	}

	secrets := []*swarm.SecretReference{
		{
			SecretName: getSwarmSecretName(PROXY_CONTAINER_NAME, "nginx.conf"),
			SecretID:   storedSecrets.nginxConfId,
			File: &swarm.SecretReferenceFileTarget{
				Name: "nginx.conf",
				UID:  "0",
				GID:  "0",
			},
		},
		{
			SecretName: getSwarmSecretName(PROXY_CONTAINER_NAME, "vchains.conf"),
			SecretID:   storedSecrets.vchainConfId,
			File: &swarm.SecretReferenceFileTarget{
				Name: "vchains.conf",
				UID:  "0",
				GID:  "0",
			},
		},
	}

	spec := getNginxServiceSpec(secrets)

	return &dockerSwarmRunner{
		client: d.client,
		spec:   spec,
	}, nil
}
