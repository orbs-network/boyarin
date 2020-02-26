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
	for serviceCfg, service := range b.config.Services().AsMap() {
		if err := b.provisionService(ctx, serviceCfg, service); err != nil {
			errors = append(errors, err)
		}
	}

	return utils.AggregateErrors(errors)
}

func (b *boyar) provisionService(ctx context.Context, cfg config.ServiceConfig, service *config.Service) error {
	if b.cache.services.CheckNewJsonValue(cfg.Name, service) {
		if service != nil {
			fullServiceName := b.config.PrefixedContainerName(cfg.Name)
			imageName := service.DockerConfig.FullImageName()

			if service.Disabled {
				return fmt.Errorf("signer service is disabled")
			}

			if service.DockerConfig.Pull {
				if err := b.orchestrator.PullImage(ctx, imageName); err != nil {
					return fmt.Errorf("could not pull docker image: %s", err)
				}
			}

			serviceConfig := &adapter.ServiceConfig{
				ImageName:     imageName,
				ContainerName: fullServiceName,
				Executable:    cfg.Executable,
				External:      cfg.External,
				HttpPort:      service.Port,

				LimitedMemory:  service.DockerConfig.Resources.Limits.Memory,
				LimitedCPU:     service.DockerConfig.Resources.Limits.CPUs,
				ReservedMemory: service.DockerConfig.Resources.Reservations.Memory,
				ReservedCPU:    service.DockerConfig.Resources.Reservations.CPUs,
			}

			jsonConfig, _ := json.Marshal(service.Config)

			var keyPairConfigJSON = getKeyConfigJson(b.config, cfg.NeedsKeys)
			appConfig := &adapter.AppConfig{
				KeyPair: keyPairConfigJSON,
				Config:  jsonConfig,
			}

			if err := b.orchestrator.RunService(ctx, serviceConfig, appConfig); err == nil {
				b.logger.Info("updated service configuration", log.Service(cfg.Name))
			} else {
				b.logger.Error("failed to update service configuration", log.Service(cfg.Name), log.Error(err))
				b.cache.services.Clear(cfg.Name)
				return err
			}
		}
	}

	return nil
}
