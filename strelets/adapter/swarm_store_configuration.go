package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/swarm"
	"strings"
)

func (d *dockerSwarm) storeConfiguration(ctx context.Context, containerName string, config *AppConfig) (*dockerSwarmSecretsConfig, error) {
	secrets := &dockerSwarmSecretsConfig{}

	if keyPairSecretId, err := d.saveSwarmSecret(ctx, containerName, "keyPair", config.KeyPair); err != nil {
		return nil, fmt.Errorf("could not store key pair secret: %s", err)
	} else {
		secrets.keysSecretId = keyPairSecretId
	}

	if networkSecretId, err := d.saveSwarmSecret(ctx, containerName, "network", config.Network); err != nil {
		return nil, fmt.Errorf("could not store network config secret: %s", err)
	} else {
		secrets.networkSecretId = networkSecretId
	}

	return secrets, nil
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
