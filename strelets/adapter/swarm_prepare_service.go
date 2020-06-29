package adapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/swarm"
)

func getVolumeNameServiceForLogs(name string) string {
	return fmt.Sprintf("%s-logs", name)
}

func (d *dockerSwarmOrchestrator) RunService(ctx context.Context, serviceConfig *ServiceConfig, appConfig *AppConfig) error {
	if err := d.RemoveService(ctx, serviceConfig.ContainerName); err != nil {
		return err
	}

	var networks []swarm.NetworkAttachmentConfig
	if serviceConfig.SignerNetworkEnabled {
		signerNetwork, err := d.getNetwork(ctx, SHARED_SIGNER_NETWORK)
		if err != nil {
			return err
		}

		networks = append(networks, signerNetwork)
	}

	if serviceConfig.ServicesNetworkEnabled {
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

	mounts, err := d.provisionStatusVolume(ctx, serviceConfig.ContainerName, ORBS_STATUS_TARGET)
	if err != nil {
		return err
	}

	if logsMount, err := d.provisionVolume(ctx, getVolumeNameServiceForLogs(serviceConfig.Name), "/opt/orbs/logs", 1, d.options); err != nil {
		return err
	} else {
		mounts = append(mounts, logsMount)
	}

	if cacheMounts, err := d.provisionCacheVolume(ctx, serviceConfig.ContainerName, ORBS_CACHE_TARGET); err != nil {
		return err
	} else {
		mounts = append(mounts, cacheMounts...)
	}

	spec := getServiceSpec(serviceConfig, secrets, networks, mounts)

	return d.create(ctx, spec, serviceConfig.ImageName)
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
			ContainerSpec: getServiceContainerSpec(serviceConfig.ImageName, serviceConfig.Executable, secrets, mounts),
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

func getServiceContainerSpec(imageName string, executable string, secrets []*swarm.SecretReference, mounts []mount.Mount) *swarm.ContainerSpec {
	subcommand := []string{
		executable,
	}

	for _, secret := range secrets {
		subcommand = append(subcommand, "--config", "/run/secrets/"+secret.File.Name)
	}

	subcommand = append(subcommand, "| multilog t s16777215 n3 '!tai64nlocal' /opt/orbs/logs 2>&1")

	command := []string{
		"/bin/sh",
		"-c",
		strings.Join(subcommand, " "),
	}

	return &swarm.ContainerSpec{
		Image:   imageName,
		Command: command,
		Secrets: secrets,
		Sysctls: GetSysctls(),
		Mounts:  mounts,
	}
}
