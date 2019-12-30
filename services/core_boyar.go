package services

import (
	"context"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/scribe/log"
	"time"
)

type BoyarService struct {
	cache   *boyar.Cache
	logger  log.Logger
	healthy bool
}

func NewCoreBoyarService(logger log.Logger) *BoyarService {
	return &BoyarService{
		cache:  boyar.NewCache(),
		logger: logger,
	}
}

func (coreBoyar *BoyarService) OnConfigChange(timeout time.Duration, cfg config.NodeConfiguration, maxDelay time.Duration) error {
	// random delay when provisioning change (that is, not bootstrap flow or repairing broken system)
	if coreBoyar.healthy {
		randomDelay(cfg, maxDelay, coreBoyar.logger)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	orchestrator, err := adapter.NewDockerSwarm(cfg.OrchestratorOptions())
	if err != nil {
		return err
	}
	defer orchestrator.Close()

	s := strelets.NewStrelets(orchestrator)
	b := boyar.NewBoyar(s, cfg, coreBoyar.cache, coreBoyar.logger)

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

func randomDelay(cfg config.NodeConfiguration, maxDelay time.Duration, logger log.Logger) {
	reloadTimeDelay := cfg.ReloadTimeDelay(maxDelay)
	logger.Info("waiting to apply new configuration", log.String("delay", maxDelay.String()))
	<-time.After(reloadTimeDelay)
}