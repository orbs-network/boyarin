package services

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"os"
	"time"
)

func Execute(ctx context.Context, flags *config.Flags, logger log.Logger) (govnr.ShutdownWaiter, error) {
	if flags.ConfigUrl == "" {
		return nil, fmt.Errorf("--config-url is a required parameter for provisioning flow")
	}

	if flags.KeyPairConfigPath == "" {
		return nil, fmt.Errorf("--keys is a required parameter for provisioning flow")
	}

	// crucial for a proper shutdown
	ctxWithCancel, cancelAndExit := context.WithCancel(ctx)
	supervisor := &govnr.TreeSupervisor{}

	// clean up old files
	logger.Info("cleaning up old files")
	os.Remove(flags.MetricsFilePath)
	os.Remove(flags.StatusFilePath)

	if flags.StatusFilePath == "" && flags.MetricsFilePath == "" {
		logger.Info("status file path and metrics file path are empty, periodical report disabled")
	} else {
		supervisor.Supervise(WatchAndReportStatusAndMetrics(ctxWithCancel, logger, flags))
	}

	cfgFetcher := NewConfigurationPollService(flags, logger)
	coreBoyar := NewCoreBoyarService(logger)

	configUpdateTimestamp := time.Now()

	// wire cfg and boyar
	supervisor.Supervise(govnr.Forever(ctxWithCancel, "apply config changes", utils.NewLogErrors("apply config changes", logger), func() {
		var cfg config.NodeConfiguration = nil
		select {
		case <-ctx.Done():
			return
		case <-time.After(flags.BootstrapResetTimeout):
		case cfg = <-cfgFetcher.Output:
		}

		if cfg == nil {
			logger.Error("unexpected empty configuration received and ignored")

			if resetInNanos := flags.BootstrapResetTimeout.Nanoseconds(); resetInNanos > 0 && time.Since(configUpdateTimestamp).Nanoseconds() >= resetInNanos {
				logger.Error(fmt.Sprintf("did not receive new valid configuratin for %s, shutting down", flags.BootstrapResetTimeout))
				cancelAndExit()
			}

			return
		}
		// random delay when provisioning change (that is, not bootstrap flow or repairing broken system)
		if coreBoyar.healthy {
			maybeDelayConfigUpdate(ctxWithCancel, cfg, flags.MaxReloadTimeDelay, coreBoyar.logger)
		} else {
			logger.Info("applying new configuration immediately")
		}

		if shouldExit := coreBoyar.CheckForUpdates(flags, cfg.OrchestratorOptions().ExecutableImage); shouldExit {
			logger.Info("shutting down after updating boyar binary")
			cancelAndExit()
			return
		}

		ctxWithTimeout, cancel := context.WithTimeout(ctxWithCancel, flags.Timeout)
		defer cancel()

		err := coreBoyar.OnConfigChange(ctxWithTimeout, cfg)
		if err != nil {
			logger.Error("error executing configuration", log.Error(err))
			cfgFetcher.Resend()
		}

		if ctxWithTimeout.Err() != nil {
			logger.Error("failed to apply new configuration", log.Error(ctxWithTimeout.Err()))
			return
		}
	}))

	supervisor.Supervise(cfgFetcher.Start(ctxWithCancel))

	return supervisor, nil
}
