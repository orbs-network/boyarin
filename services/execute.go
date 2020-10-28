package services

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"os"
)

func Execute(ctx context.Context, flags *config.Flags, logger log.Logger) (govnr.ShutdownWaiter, error) {
	if flags.ConfigUrl == "" {
		return nil, fmt.Errorf("--config-url is a required parameter for provisioning flow")
	}

	if flags.KeyPairConfigPath == "" {
		return nil, fmt.Errorf("--keys is a required parameter for provisioning flow")
	}

	supervisor := &govnr.TreeSupervisor{}

	// clean up old files
	logger.Info("cleaning up old files")
	os.Remove(flags.MetricsFilePath)
	os.Remove(flags.StatusFilePath)

	if flags.StatusFilePath == "" && flags.MetricsFilePath == "" {
		logger.Info("status file path and metrics file path are empty, periodical report disabled")
	} else {
		supervisor.Supervise(WatchAndReportStatusAndMetrics(ctx, logger, flags.StatusFilePath, flags.MetricsFilePath))
	}

	cfgFetcher := NewConfigurationPollService(flags, logger)
	coreBoyar := NewCoreBoyarService(logger)
	// for self update E2E
	if flags.TargetPath != "" {
		coreBoyar.binaryTargetPath = flags.TargetPath
	}

	// wire cfg and boyar
	supervisor.Supervise(govnr.Forever(ctx, "apply config changes", utils.NewLogErrors("apply config changes", logger), func() {
		var cfg config.NodeConfiguration = nil
		select {
		case <-ctx.Done():
			return
		case cfg = <-cfgFetcher.Output:
		}
		if cfg == nil {
			return
		}
		// random delay when provisioning change (that is, not bootstrap flow or repairing broken system)
		if coreBoyar.healthy {
			maybeDelayConfigUpdate(ctx, cfg, flags.MaxReloadTimeDelay, coreBoyar.logger)
		} else {
			logger.Info("applying new configuration immediately")
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
