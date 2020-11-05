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

	if flags.BootstrapResetTimeout > 0 && flags.BootstrapResetTimeout.Nanoseconds() <= flags.PollingInterval.Nanoseconds() {
		return nil, fmt.Errorf("invalid configuration: bootstrap reset timeout is less or equal to config polling interval")
	}

	if flags.StatusFilePath == "" && flags.MetricsFilePath == "" {
		logger.Info("status file path and metrics file path are empty, periodical report disabled")
	} else {
		supervisor.Supervise(WatchAndReportStatusAndMetrics(ctxWithCancel, logger, flags))
	}

	coreBoyar := NewCoreBoyarService(logger)
	configCache := utils.NewCacheFilter()

	configUpdateTimestamp := time.Now()

	// wire cfg and boyar
	supervisor.Supervise(govnr.Forever(ctxWithCancel, "apply config changes", utils.NewLogErrors("apply config changes", logger), func() {
		var cfg config.NodeConfiguration = nil
		var err error

		select {
		case <-ctx.Done():
			return
		case <-timeout(flags.BootstrapResetTimeout):
			logger.Error("bootstrap reset timeout reached", log.String("configUpdateTimestamp", configUpdateTimestamp.Format(time.RFC3339)))
		case <-time.After(flags.PollingInterval):
			cfg, err = config.GetConfiguration(flags)
			if err != nil {
				logger.Error("invalid configuration", log.Error(err))
			} else {
				configUpdateTimestamp = time.Now()
				logger.Info("last valid configuration timestamp updated", log.String("configUpdateTimestamp", configUpdateTimestamp.Format(time.RFC3339)))
			}
		}

		if cfg == nil {
			if resetInNanos := flags.BootstrapResetTimeout.Nanoseconds(); resetInNanos > 0 && time.Since(configUpdateTimestamp).Nanoseconds() >= resetInNanos {
				logger.Error(fmt.Sprintf("did not receive new valid configuratin for %s, shutting down", flags.BootstrapResetTimeout))
				cancelAndExit()
			}

			return
		}

		if !configCache.CheckNewValue(cfg) {
			logger.Info("configuration has not changed")
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

		err = coreBoyar.OnConfigChange(ctxWithTimeout, cfg)
		if err != nil {
			logger.Error("error executing configuration", log.Error(err))
			configCache.Clear()
		}

		if ctxWithTimeout.Err() != nil {
			logger.Error("failed to apply new configuration", log.Error(ctxWithTimeout.Err()))
			return
		}
	}))

	return supervisor, nil
}

func timeout(duration time.Duration) <-chan time.Time {
	if duration == 0 {
		return make(chan time.Time) // empty channel that nobody for waiting forever
	}

	return time.After(duration)
}
