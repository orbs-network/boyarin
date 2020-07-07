package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
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
	if b.cache.services.CheckNewJsonValue(serviceName, service) {
		if service != nil {
			fullServiceName := b.config.NamespacedContainerName(serviceName)
			imageName := service.DockerConfig.FullImageName()

			if service.Disabled {
				return fmt.Errorf("service %s is disabled even though it should not be, ignored", serviceName)
			}

			if service.DockerConfig.Pull {
				if err := b.orchestrator.PullImage(ctx, imageName); err != nil {
					return fmt.Errorf("could not pull docker image: %s", err)
				}
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
			}

			jsonConfig, _ := json.Marshal(service.Config)

			var keyPairConfigJSON = getKeyConfigJson(b.config, !service.InjectNodePrivateKey)
			appConfig := &adapter.AppConfig{
				KeyPair: keyPairConfigJSON,
				Config:  jsonConfig,
			}

			if err := b.orchestrator.RunService(ctx, serviceConfig, appConfig); err == nil {
				b.logger.Info("updated service configuration", log.Service(serviceName))
			} else {
				b.logger.Error("failed to update service configuration", log.Service(serviceName), log.Error(err))
				b.cache.services.Clear(serviceName)
				return err
			}
		}
	}

	return nil
}
