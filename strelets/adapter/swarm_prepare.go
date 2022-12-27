package adapter

import (
	"context"
	"github.com/docker/docker/api/types/swarm"
)

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

func getServiceMode(replicas uint64) swarm.ServiceMode {
	return swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{
			Replicas: &replicas,
		},
	}
}

const MEGABYTE = 1024 * 1024
const CPU_SHARES = 1000000000

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
		MemoryBytes: 3000 * MEGABYTE,
		NanoCPUs:    1 * CPU_SHARES,
	}, limitMemory, limitCPU)
	reservations := overrideResource(&swarm.Resources{}, reserveMemory, reserveCPU)

	return &swarm.ResourceRequirements{
		Limits:       limits,
		Reservations: reservations,
	}
}

func (d *dockerSwarmOrchestrator) getNetwork(ctx context.Context, name string) (network swarm.NetworkAttachmentConfig, err error) {
	target, err := d.GetOverlayNetwork(ctx, name)
	if err != nil {
		return swarm.NetworkAttachmentConfig{}, err
	}

	return swarm.NetworkAttachmentConfig{
		Target: target,
	}, nil
}
