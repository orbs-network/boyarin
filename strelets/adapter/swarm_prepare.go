package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"time"
)

func (d *dockerSwarm) Prepare(ctx context.Context, imageName string, containerName string, httpPort int, gossipPort int, appConfig *AppConfig) (Runner, error) {
	return &dockerSwarmRunner{
		client: d.client,
		spec: func() (swarm.ServiceSpec, error) {
			config, err := d.storeVirtualChainConfiguration(ctx, containerName, appConfig)
			if err != nil {
				return swarm.ServiceSpec{}, err
			}

			secrets := []*swarm.SecretReference{
				getSecretReference(containerName, config.configSecretId, "config", "config.json"),
				getSecretReference(containerName, config.keysSecretId, "keyPair", "keys.json"),
				getSecretReference(containerName, config.networkSecretId, "network", "network.json"),
			}

			mounts, err := d.provisionVolumes(ctx, containerName)
			if err != nil {
				return swarm.ServiceSpec{}, fmt.Errorf("failed to provision volumes: %s", err)
			}

			return getVirtualChainServiceSpec(imageName, containerName, httpPort, gossipPort, secrets, mounts), nil
		},
		serviceName: getServiceId(containerName),
		imageName:   imageName,
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

func getContainerSpec(imageName string, secrets []*swarm.SecretReference, mounts []mount.Mount) *swarm.ContainerSpec {
	command := []string{
		"/opt/orbs/orbs-node",
		"--silent",
		"--log", "/opt/orbs/logs/node.log",
	}

	for _, secret := range secrets {
		command = append(command, "--config", "/var/run/secrets/"+secret.File.Name)
	}

	return &swarm.ContainerSpec{
		Image:   imageName,
		Command: command,
		Secrets: secrets,
		Sysctls: getSysctls(),
		Mounts:  mounts,
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

const MEGABYTE = 1024
const CPU_SHARES = 10000000000

func overrideResource(resource *swarm.Resources, memory int64, cpu float64) *swarm.Resources {
	if memory != 0 {
		resource.MemoryBytes = memory * MEGABYTE
	}

	if cpu != 0 {
		resource.NanoCPUs = int64(cpu * CPU_SHARES)
	}

	return resource
}

func getResourceRequirements(limitMemory int64, limitCPU float64, reserveMemory int64, reserveCPU float64) *swarm.ResourceRequirements {
	limits := overrideResource(&swarm.Resources{
		MemoryBytes: 2000 * MEGABYTE,
		NanoCPUs:    1 * CPU_SHARES,
	}, limitMemory, limitCPU)

	reservations := overrideResource(&swarm.Resources{
		MemoryBytes: 200 * MEGABYTE,
		NanoCPUs:    0.25 * CPU_SHARES,
	}, reserveMemory, reserveCPU)

	return &swarm.ResourceRequirements{
		Limits:       limits,
		Reservations: reservations,
	}
}

func getVirtualChainServiceSpec(imageName string, containerName string, httpPort int, gossipPort int, secrets []*swarm.SecretReference, mounts []mount.Mount) swarm.ServiceSpec {
	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: getContainerSpec(imageName, secrets, mounts),
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
			//Resources: getResourceRequirements(),
		},
		Mode:         getServiceMode(replicas),
		EndpointSpec: getEndpointsSpec(httpPort, gossipPort),
	}
	spec.Name = getServiceId(containerName)

	return spec
}
