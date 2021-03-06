package adapter

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
	"time"
)

func (d *dockerSwarmOrchestrator) RunService(ctx context.Context, serviceConfig *ServiceConfig, appConfig *AppConfig) error {
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

	if serviceConfig.AllowAccessToServices {
		servicesNetwork, err := d.getNetwork(ctx, SHARED_SERVICES_NETWORK)
		if err != nil {
			return err
		}

		networks = append(networks, servicesNetwork)
	}

	config, err := d.storeServiceConfiguration(ctx, serviceConfig.ContainerName, appConfig)
	if err != nil {
		return err
	}

	secrets := []*swarm.SecretReference{
		getSecretReference(serviceConfig.ContainerName, config.configSecretId, "config", "config.json"),
	}
	if config.keysSecretId != "" {
		secrets = append(secrets, getSecretReference(serviceConfig.ContainerName, config.keysSecretId, "keyPair", "keys.json"))
	}

	mounts, err := d.provisionServiceVolumes(ctx, serviceConfig.ContainerName, serviceConfig.LogsMountPointNames)
	if err != nil {
		return err
	}

	spec := getServiceSpec(serviceConfig, secrets, networks, mounts)

	return d.create(ctx, spec, serviceConfig.ImageName)
}

func (d *dockerSwarmOrchestrator) provisionServiceVolumes(ctx context.Context, containerName string, logsMountPointNames map[string]string) (mounts []mount.Mount, err error) {
	if statusMount, err := d.provisionStatusVolume(ctx, containerName, ORBS_STATUS_TARGET); err != nil {
		return nil, err
	} else {
		mounts = append(mounts, statusMount)
	}

	if cacheMount, err := d.provisionCacheVolume(ctx, containerName); err != nil {
		return nil, err
	} else {
		mounts = append(mounts, cacheMount)
	}

	if len(logsMountPointNames) == 0 {
		if logsMount, err := d.provisionLogsVolume(ctx, containerName, ORBS_LOGS_TARGET); err != nil {
			return nil, fmt.Errorf("failed to provision volumes: %s", err)
		} else {
			mounts = append(mounts, logsMount)
		}
	} else {
		// special case for multiple logs
		for simpleName, namespacedName := range logsMountPointNames {
			if logsMount, err := d.provisionLogsVolume(ctx, namespacedName, GetNestedLogsMountPath(simpleName)); err != nil {
				return nil, fmt.Errorf("failed to provision volumes: %s", err)
			} else {
				mounts = append(mounts, logsMount)
			}
		}
	}

	return mounts, nil
}

func (d *dockerSwarmOrchestrator) storeServiceConfiguration(ctx context.Context, containerName string, config *AppConfig) (*dockerSwarmSecretsConfig, error) {
	secrets := &dockerSwarmSecretsConfig{}

	if configSecretId, err := d.saveSwarmSecret(ctx, containerName, "config", config.Config); err != nil {
		return nil, fmt.Errorf("could not store config secret: %s", err)
	} else {
		secrets.configSecretId = configSecretId
	}

	if config.KeyPair != nil {
		if keyPairSecretId, err := d.saveSwarmSecret(ctx, containerName, "keyPair", config.KeyPair); err != nil {
			return nil, fmt.Errorf("could not store key pair secret: %s", err)
		} else {
			secrets.keysSecretId = keyPairSecretId
		}
	}

	return secrets, nil
}

func getServiceSpec(serviceConfig *ServiceConfig, secrets []*swarm.SecretReference, networks []swarm.NetworkAttachmentConfig, mounts []mount.Mount) swarm.ServiceSpec {
	restartDelay := time.Duration(10 * time.Second)
	replicas := uint64(1)

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: getServiceContainerSpec(serviceConfig.ImageName, serviceConfig.ExecutablePath, secrets, mounts),
			RestartPolicy: &swarm.RestartPolicy{
				Delay: &restartDelay,
			},
			Resources: getResourceRequirements(serviceConfig.LimitedMemory, serviceConfig.LimitedCPU,
				serviceConfig.ReservedMemory, serviceConfig.ReservedCPU),
		},
		Networks: networks,
		Mode:     getServiceMode(replicas),
	}

	if serviceConfig.ExternalPort != 0 {
		spec.EndpointSpec = &swarm.EndpointSpec{
			Ports: []swarm.PortConfig{
				{
					Protocol:      "tcp",
					PublishMode:   swarm.PortConfigPublishModeIngress,
					PublishedPort: uint32(serviceConfig.ExternalPort),
					TargetPort:    uint32(serviceConfig.InternalPort),
				},
			},
		}
	}

	spec.Name = serviceConfig.ContainerName

	return spec
}

const SERVICE_EXECUTABLE_PATH = "/opt/orbs/service"

func getServiceContainerSpec(imageName string, executable string, secrets []*swarm.SecretReference, mounts []mount.Mount) *swarm.ContainerSpec {
	if executable == "" {
		executable = SERVICE_EXECUTABLE_PATH
	}

	command := []string{
		executable,
	}

	for _, secret := range secrets {
		command = append(command, "--config", "/run/secrets/"+secret.File.Name)
	}

	return &swarm.ContainerSpec{
		Image:   imageName,
		Command: command,
		Secrets: secrets,
		Sysctls: GetSysctls(),
		Mounts:  mounts,
	}
}
