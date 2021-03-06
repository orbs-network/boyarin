package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"time"
)

var VIRTUAL_CHAIN_RESTART_DELAY = 20 * time.Minute
var VIRTUAL_CHAIN_RESTART_SUCCESS_WINDOW = 2 * time.Minute

func (d *dockerSwarmOrchestrator) RunVirtualChain(ctx context.Context, serviceConfig *ServiceConfig, appConfig *AppConfig) error {
	if err := d.RemoveService(ctx, serviceConfig.ContainerName); err != nil {
		return err
	}

	var networks []swarm.NetworkAttachmentConfig
	if serviceConfig.AllowAccessToSigner {
		signerNetwork, err := d.getNetwork(ctx, SHARED_SIGNER_NETWORK)
		if err != nil {
			return err
		}

		networks = append(networks, signerNetwork)
	}

	if serviceConfig.HTTPProxyNetworkEnabled {
		proxyNetwork, err := d.getNetwork(ctx, SHARED_PROXY_NETWORK)
		if err != nil {
			return err
		}

		networks = append(networks, proxyNetwork)
	}

	if serviceConfig.AllowAccessToServices {
		servicesNetwork, err := d.getNetwork(ctx, SHARED_SERVICES_NETWORK)
		if err != nil {
			return err
		}

		networks = append(networks, servicesNetwork)
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

	mounts, err := d.provisionServiceVolumes(ctx, serviceConfig.ContainerName, nil)
	if err != nil {
		return err
	}

	blocksMount, err := d.provisionVchainVolume(ctx, serviceConfig.NodeAddress, serviceConfig.Id)
	if err != nil {
		return fmt.Errorf("failed to provision volumes: %s", err)
	} else {
		mounts = append(mounts, blocksMount)
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

func getVirtualChainServiceSpec(serviceConfig *ServiceConfig, secrets []*swarm.SecretReference, mounts []mount.Mount, networks []swarm.NetworkAttachmentConfig) swarm.ServiceSpec {
	replicas := uint64(1)

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: getServiceContainerSpec(serviceConfig.ImageName, serviceConfig.ExecutablePath, secrets, mounts),
			RestartPolicy: &swarm.RestartPolicy{
				Delay:     &VIRTUAL_CHAIN_RESTART_DELAY,
				Window:    &VIRTUAL_CHAIN_RESTART_SUCCESS_WINDOW,
				Condition: swarm.RestartPolicyConditionOnFailure,
			},
			Resources: getResourceRequirements(serviceConfig.LimitedMemory, serviceConfig.LimitedCPU,
				serviceConfig.ReservedMemory, serviceConfig.ReservedCPU),
		},
		Networks: networks,
		Mode:     getServiceMode(replicas),
		EndpointSpec: &swarm.EndpointSpec{
			Ports: []swarm.PortConfig{
				{
					Protocol:      "tcp",
					PublishMode:   swarm.PortConfigPublishModeIngress,
					PublishedPort: uint32(serviceConfig.ExternalPort),
					TargetPort:    uint32(serviceConfig.InternalPort),
				},
			},
		},
	}
	spec.Name = serviceConfig.ContainerName

	return spec
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
