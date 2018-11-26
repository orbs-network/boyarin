package adapter

import (
	"context"
	"github.com/docker/docker/api/types/swarm"
	"time"
)

func (d *dockerSwarm) Prepare(ctx context.Context, imageName string, containerName string, httpPort int, gossipPort int, appConfig *AppConfig) (Runner, error) {
	config, err := d.storeConfiguration(ctx, containerName, appConfig)
	if err != nil {
		return nil, err
	}

	secrets := []*swarm.SecretReference{
		getSecretReference(containerName, config.keysSecretId, "keyPair", "keys.json"),
		getSecretReference(containerName, config.networkSecretId, "network", "network.json"),
	}

	spec := getServiceSpec(imageName, containerName, httpPort, gossipPort, secrets)

	return &dockerSwarmRunner{
		client: d.client,
		spec:   spec,
	}, nil
}

func getSecretReference(containerName string, secretId string, secretName string, filename string) *swarm.SecretReference {
	return &swarm.SecretReference{
		SecretName: getSwarmSecretName(containerName, secretName),
		SecretID:   secretId,
		File: &swarm.SecretReferenceFileTarget{
			Name: filename,
			UID:  "0",
			GID:  "0",
		},
	}
}

func getContainerSpec(imageName string, secrets []*swarm.SecretReference) *swarm.ContainerSpec {
	command := []string{
		"/opt/orbs/orbs-node",
		"--silent",
		// FIXME add separate volume for logs
		"--log", "/opt/orbs/node.log",
	}

	for _, secret := range secrets {
		command = append(command, "--config", "/var/run/secrets/"+secret.File.Name)
	}

	return &swarm.ContainerSpec{
		Image:   imageName,
		Command: command,
		Secrets: secrets,
	}
}

func getServiceMode(replicas uint64) swarm.ServiceMode {
	return swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{
			Replicas: &replicas,
		},
	}
}

func getEndpointsSpec(httpPort int, gossipPort int) *swarm.EndpointSpec {
	return &swarm.EndpointSpec{
		Ports: []swarm.PortConfig{
			{
				Protocol:      "tcp",
				PublishMode:   swarm.PortConfigPublishModeIngress,
				PublishedPort: uint32(httpPort),
				TargetPort:    8080,
			},
			{
				Protocol:      "tcp",
				PublishMode:   swarm.PortConfigPublishModeIngress,
				PublishedPort: uint32(gossipPort),
				TargetPort:    4400,
			},
		},
	}
}

func getServiceSpec(imageName string, containerName string, httpPort int, gossipPort int, secrets []*swarm.SecretReference) swarm.ServiceSpec {
	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: getContainerSpec(imageName, secrets),
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
		},
		Mode:         getServiceMode(replicas),
		EndpointSpec: getEndpointsSpec(httpPort, gossipPort),
	}
	spec.Name = getServiceId(containerName)

	return spec
}
