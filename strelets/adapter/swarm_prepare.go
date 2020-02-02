package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"time"
)

func (d *dockerSwarmOrchestrator) RunVirtualChain(ctx context.Context, serviceConfig *ServiceConfig, appConfig *AppConfig) error {
	serviceName := GetServiceId(serviceConfig.ContainerName)

	if err := d.ServiceRemove(ctx, serviceName); err != nil {
		return err
	}

	networks, err := d.getNetworks(ctx, SHARED_SIGNER_NETWORK)
	if err != nil {
		return err
	}

	config, err := d.storeVirtualChainConfiguration(ctx, serviceConfig.ContainerName, appConfig)
	if err != nil {
		return err
	}

	secrets := []*swarm.SecretReference{
		getSecretReference(serviceConfig.ContainerName, config.configSecretId, "config", "config.json"),
		getSecretReference(serviceConfig.ContainerName, config.keysSecretId, "keyPair", "keys.json"),
		getSecretReference(serviceConfig.ContainerName, config.networkSecretId, "network", "network.json"),
	}

	mounts, err := d.provisionVolumes(ctx, serviceConfig.NodeAddress, serviceConfig.Id,
		defaultValue(serviceConfig.BlocksVolumeSize, 100), defaultValue(serviceConfig.LogsVolumeSize, 2))
	if err != nil {
		return fmt.Errorf("failed to provision volumes: %s", err)
	}

	spec := getVirtualChainServiceSpec(serviceConfig, secrets, mounts, networks)

	return d.create(ctx, spec, serviceConfig.ImageName)
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
		Sysctls: GetSysctls(),
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
	limits := overrideResource(&swarm.Resources{}, limitMemory, limitCPU)
	reservations := overrideResource(&swarm.Resources{}, reserveMemory, reserveCPU)

	return &swarm.ResourceRequirements{
		Limits:       limits,
		Reservations: reservations,
	}
}

func getVirtualChainServiceSpec(serviceConfig *ServiceConfig, secrets []*swarm.SecretReference, mounts []mount.Mount, networks []swarm.NetworkAttachmentConfig) swarm.ServiceSpec {
	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: getContainerSpec(serviceConfig.ImageName, secrets, mounts),
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
			Resources: getResourceRequirements(serviceConfig.LimitedMemory, serviceConfig.LimitedCPU,
				serviceConfig.ReservedMemory, serviceConfig.ReservedCPU),
		},
		Networks:     networks,
		Mode:         getServiceMode(replicas),
		EndpointSpec: getEndpointsSpec(serviceConfig.HttpPort, serviceConfig.GossipPort),
	}
	spec.Name = GetServiceId(serviceConfig.ContainerName)

	return spec
}

func (d *dockerSwarmOrchestrator) getNetworks(ctx context.Context, name string) (networks []swarm.NetworkAttachmentConfig, err error) {
	target, err := d.GetOverlayNetwork(ctx, name)
	if err != nil {
		return nil, err
	}

	networks = append(networks, swarm.NetworkAttachmentConfig{
		Target: target,
	})

	return
}

func defaultValue(value, defaultV int) int {
	if value != 0 {
		return value
	}

	return defaultV
}
