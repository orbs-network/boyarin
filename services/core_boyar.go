package services

import (
	"context"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/scribe/log"
	"time"
)

type BoyarService struct {
	cache            *boyar.Cache
	logger           log.Logger
	healthy          bool
	binaryTargetPath string
}

func NewCoreBoyarService(logger log.Logger) *BoyarService {
	return &BoyarService{
		cache:  boyar.NewCache(),
		logger: logger,
	}
}

func (coreBoyar *BoyarService) OnConfigChange(ctx context.Context, cfg config.NodeConfiguration) error {

	orchestrator, err := adapter.NewDockerSwarm(cfg.OrchestratorOptions(), coreBoyar.logger)
	if err != nil {
		return err
	}
	defer orchestrator.Close()

	b := boyar.NewBoyar(orchestrator, cfg, coreBoyar.cache, coreBoyar.logger)

	var errors []error

	if err := b.ProvisionServices(ctx); err != nil {
		errors = append(errors, err)
	}

	if err := b.ProvisionVirtualChains(ctx); err != nil {
		errors = append(errors, err)
	}

	if err := b.ProvisionHttpAPIEndpoint(ctx); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		coreBoyar.healthy = false
		return utils.AggregateErrors(errors)
	}

	coreBoyar.healthy = true
	return nil
}

func maybeDelayConfigUpdate(ctx context.Context, cfg config.NodeConfiguration, maxDelay time.Duration, logger log.Logger) {
	reloadTimeDelay := cfg.ReloadTimeDelay(maxDelay)
	if reloadTimeDelay.Seconds() > 1 { // the delay is designed to break symmetry between nodes. less than a second is practically zero
		logger.Info("waiting to update configuration", log.String("delay", maxDelay.String()))
		select {
		case <-time.After(reloadTimeDelay):
		case <-ctx.Done():
		}
	} else {
		logger.Info("updating configuration immediately")
	}
}
