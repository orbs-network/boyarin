package services

import (
	"context"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/scribe/log"
)

func Bootstrap(ctx context.Context, cfg config.NodeConfiguration, logger log.Logger) error {
	logger.Info("bootstrapping from default config")

	coreBoyar := NewCoreBoyarService(logger)
	return coreBoyar.OnConfigChange(ctx, cfg)
}
