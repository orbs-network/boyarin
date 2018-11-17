package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"io"
	"os"
	"strings"
	"time"
)

type dockerSwarm struct {
	client *client.Client
}

type dockerSwarmSecretsConfig struct {
	networkSecretId string
	keysSecretId    string
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
	config := dockerConfig.(*dockerSwarmSecretsConfig)

	ureplicas := uint64(1)
	restartDelay := time.Duration(10 * time.Second)

	fmt.Println(config)

	keysSecret := &swarm.SecretReference{
		SecretName: getSwarmSecretName(containerName, "keyPair"),
		SecretID:   config.keysSecretId,
		File: &swarm.SecretReferenceFileTarget{
			Name: "keys.json",
			UID:  "0",
			GID:  "0",
		},
	}

	networkSecret := &swarm.SecretReference{
		SecretName: getSwarmSecretName(containerName, "network"),
		SecretID:   config.networkSecretId,
		File: &swarm.SecretReferenceFileTarget{
			Name: "network.json",
			UID:  "0",
			GID:  "0",
		},
	}

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Command: []string{"top"},
				Image:   "orbs:export",
				//Command: config.Cmd,
				Secrets: []*swarm.SecretReference{
					keysSecret,
					networkSecret,
				},
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

func (d *dockerSwarm) StoreConfiguration(ctx context.Context, containerName string, root string, config *AppConfig) (interface{}, error) {
	secrets := &dockerSwarmSecretsConfig{}

	if keyPairSecretId, err := d.saveSwarmSecret(ctx, containerName, "keyPair", config.KeyPair); err != nil {
		return nil, err
	} else {
		secrets.keysSecretId = keyPairSecretId
	}

	if networkSecretId, err := d.saveSwarmSecret(ctx, containerName, "network", config.Network); err != nil {
		return nil, err
	} else {
		secrets.networkSecretId = networkSecretId
	}

	return secrets, nil
}

func (d *dockerSwarm) GetContainerConfiguration(imageName string, containerName string, root string, httpPort int, gossipPort int, storedConfig interface{}) interface{} {
	// FIXME return proper struct
	return storedConfig
}

func getServiceId(input string) string {
	return "stack-" + input
}

func (d *dockerSwarm) saveSwarmSecret(ctx context.Context, containerName string, secretName string, content []byte) (string, error) {
	secretId := getSwarmSecretName(containerName, secretName)
	d.client.SecretRemove(ctx, secretId)

	secretSpec := swarm.SecretSpec{
		Data: content,
	}
	secretSpec.Name = secretId

	response, err := d.client.SecretCreate(ctx, secretSpec)
	return response.ID, err
}

func getSwarmSecretName(containerName string, secretName string) string {
	return strings.Join([]string{containerName, secretName}, "-")
}
