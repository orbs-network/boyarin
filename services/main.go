package services

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
)

func Execute(ctx context.Context, flags *config.Flags, logger log.Logger) (govnr.ShutdownWaiter, error) {
	if flags.ConfigUrl == "" {
		return nil, fmt.Errorf("--config-url is a required parameter for provisioning flow")
	}

	if flags.KeyPairConfigPath == "" {
		return nil, fmt.Errorf("--keys is a required parameter for provisioning flow")
	}

	supervisor := &govnr.TreeSupervisor{}
	supervisor.Supervise(WatchAndReportServicesStatus(ctx, logger))

	cfgFetcher := NewConfigurationPollService(flags, logger)
	coreBoyar := NewCoreBoyarService(logger)

	// wire cfg and boyar
	supervisor.Supervise(govnr.Forever(ctx, "apply config changes", utils.NewLogErrors("apply config changes", logger), func() {
		cfg := <-cfgFetcher.Output

		// random delay when provisioning change (that is, not bootstrap flow or repairing broken system)
		if coreBoyar.healthy {
			randomDelay(ctx, cfg, flags.MaxReloadTimeDelay, coreBoyar.logger)
		}

		ctx, cancel := context.WithTimeout(ctx, flags.Timeout)
		defer cancel()
		if ctx.Err() != nil {
			return
		}
		err := coreBoyar.OnConfigChange(ctx, cfg)
		if err != nil {
			logger.Error("error executing configuration", log.Error(err))
			cfgFetcher.Resend()
		}
	}))

	supervisor.Supervise(cfgFetcher.Start(ctx))

	return supervisor, nil
}
