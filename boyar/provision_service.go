package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/pkg/errors"
)

func (b *boyar) ProvisionServices(ctx context.Context) error {
	if _, err := b.orchestrator.GetOverlayNetwork(ctx, adapter.SHARED_SIGNER_NETWORK); err != nil {
		return errors.Wrap(err, "failed creating network")
	}

	var errors []error
	for serviceName, service := range b.config.Services() {
		if err := b.provisionService(ctx, serviceName, service); err != nil {
			errors = append(errors, err)
		}
	}

	return utils.AggregateErrors(errors)
}

func (b *boyar) provisionService(ctx context.Context, serviceName string, service *config.Service) error {
	if service == nil {
		return nil
	}

	logger := b.logger.WithTags(log.String("service", serviceName))

	fullServiceName := b.config.NamespacedContainerName(serviceName)
	imageName := service.DockerConfig.FullImageName()

	if service.Disabled {
		if b.cache.services.CheckNewJsonValue(serviceName, removed) {
			if err := b.orchestrator.RemoveService(ctx, serviceName); err != nil {
				b.cache.services.Clear(serviceName)
				logger.Error("failed to remove service", log.Error(err))
			} else {
				if service.PurgeData {
					govnr.Once(utils.NewLogErrors("apply config changes", logger), func() {
						// FIXME add purge data to the orchestrator
						//err := b.orchestrator.PurgeData()
					})
				}

				logger.Info("successfully removed service")
			}
		}

		return nil
	}

	var logsMountPointNames map[string]string
	if service.MountNodeLogs {
		logsMountPointNames = getLogsMountPointNames(b.config)
	}

	serviceConfig := &adapter.ServiceConfig{
		NodeAddress: string(b.config.NodeAddress()),

		ImageName:      imageName,
		Name:           serviceName,
		ContainerName:  fullServiceName,
		ExecutablePath: service.ExecutablePath,
		InternalPort:   service.InternalPort,
		ExternalPort:   service.ExternalPort,

		AllowAccessToSigner:   service.AllowAccessToSigner,
		AllowAccessToServices: service.AllowAccessToServices,

		LimitedMemory:  service.DockerConfig.Resources.Limits.Memory,
		LimitedCPU:     service.DockerConfig.Resources.Limits.CPUs,
		ReservedMemory: service.DockerConfig.Resources.Reservations.Memory,
		ReservedCPU:    service.DockerConfig.Resources.Reservations.CPUs,

		LogsMountPointNames: logsMountPointNames,
	}

	jsonConfig, _ := json.Marshal(service.Config)

	var keyPairConfigJSON = getKeyConfigJson(b.config, !service.InjectNodePrivateKey)
	appConfig := &adapter.AppConfig{
		KeyPair: keyPairConfigJSON,
		Config:  jsonConfig,
	}

	if b.cache.services.CheckNewJsonValue(serviceName, serviceConfig) {
		if service.DockerConfig.Pull {
			if err := b.orchestrator.PullImage(ctx, imageName); err != nil {
				return fmt.Errorf("could not pull docker image: %s", err)
			}
		}

		if err := b.orchestrator.RunService(ctx, serviceConfig, appConfig); err == nil {
			data, _ := json.Marshal(serviceConfig)
			logger.Info("updated service configuration", log.String("configuration", string(data)))
		} else {
			logger.Error("failed to update service configuration", log.Error(err))
			b.cache.services.Clear(serviceName)
			return err
		}
	}

	return nil
}

func getLogsMountPointNames(cfg config.NodeConfiguration) map[string]string {
	mountPointNames := make(map[string]string)
	mountPointNames["boyar"] = "boyar"

	for name, _ := range cfg.Services() {
		mountPointNames[name] = cfg.NamespacedContainerName(name)
	}

	for _, chain := range cfg.Chains() {
		mountPointNames[chain.GetContainerName()] = cfg.NamespacedContainerName(chain.GetContainerName())
	}

	return mountPointNames
}
