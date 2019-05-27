package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/swarm"
	"time"
)

func (d *dockerSwarm) PrepareService(ctx context.Context, serviceConfig *ServiceConfig, appConfig *AppConfig) (Runner, error) {
	serviceName := GetServiceId(serviceConfig.ContainerName)

	if err := d.RemoveContainer(ctx, serviceName); err != nil {
		return nil, err
	}

	return &dockerSwarmRunner{
		client: d.client,
		spec: func() (swarm.ServiceSpec, error) {
			config, err := d.storeServiceConfiguration(ctx, serviceConfig.ContainerName, appConfig)
			if err != nil {
				return swarm.ServiceSpec{}, err
			}

			secrets := []*swarm.SecretReference{
				getSecretReference(serviceConfig.ContainerName, config.configSecretId, "config", "config.json"),
				getSecretReference(serviceConfig.ContainerName, config.keysSecretId, "keyPair", "keys.json"),
			}

			return getServiceSpec(serviceConfig, secrets), nil
		},
		serviceName: GetServiceId(serviceConfig.ContainerName),
		imageName:   serviceConfig.ImageName,
	}, nil
}

func (d *dockerSwarm) storeServiceConfiguration(ctx context.Context, containerName string, config *AppConfig) (*dockerSwarmSecretsConfig, error) {
	secrets := &dockerSwarmSecretsConfig{}

	if configSecretId, err := d.saveSwarmSecret(ctx, containerName, "config", config.Config); err != nil {
		return nil, fmt.Errorf("could not store config secret: %s", err)
	} else {
		secrets.configSecretId = configSecretId
	}

	if keyPairSecretId, err := d.saveSwarmSecret(ctx, containerName, "keyPair", config.KeyPair); err != nil {
		return nil, fmt.Errorf("could not store key pair secret: %s", err)
	} else {
		secrets.keysSecretId = keyPairSecretId
	}

	return secrets, nil
}

func getServiceSpec(serviceConfig *ServiceConfig, secrets []*swarm.SecretReference) swarm.ServiceSpec {
	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: getServiceContainerSpec(serviceConfig.ImageName, secrets),
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
			Resources: getResourceRequirements(serviceConfig.LimitedMemory, serviceConfig.LimitedCPU,
				serviceConfig.ReservedMemory, serviceConfig.ReservedCPU),
		},
		Mode: getServiceMode(replicas),
		//EndpointSpec: &swarm.EndpointSpec{
		//	Ports: []swarm.PortConfig{
		//		{
		//			Protocol: "tcp",
		//			// FIXME should publish only for overlay network
		//			PublishMode:   swarm.PortConfigPublishModeIngress,
		//			PublishedPort: uint32(serviceConfig.HttpPort),
		//			TargetPort:    uint32(serviceConfig.HttpPort),
		//		},
		//	},
		//},
	}
	spec.Name = GetServiceId(serviceConfig.ContainerName)

	return spec
}

func getServiceContainerSpec(imageName string, secrets []*swarm.SecretReference) *swarm.ContainerSpec {
	command := []string{
		"/opt/orbs/orbs-signer",
	}

	for _, secret := range secrets {
		command = append(command, "--config", "/run/secrets/"+secret.File.Name)
	}

	return &swarm.ContainerSpec{
		Image:   imageName,
		Command: command,
		Secrets: secrets,
		Sysctls: GetSysctls(),
	}
}
