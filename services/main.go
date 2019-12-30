package services

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"time"
)

func Execute(ctx context.Context, flags *config.Flags, logger log.Logger) error {
	if flags.ConfigUrl == "" {
		return fmt.Errorf("--config-url is a required parameter for provisioning flow")
	}

	if flags.KeyPairConfigPath == "" {
		return fmt.Errorf("--keys is a required parameter for provisioning flow")
	}

	supervisor := &govnr.TreeSupervisor{}
	supervisor.Supervise(WatchAndReportServicesStatus(ctx, logger))

	cfgFetcher := NewConfigurationPollService(flags, logger)
	coreBoyar := NewCoreBoyarService(logger)

	// wire cfg and boyar
	supervisor.Supervise(govnr.Forever(ctx, "apply config changes", utils.NewLogErrors(logger), func() {
		cfg := <-cfgFetcher.Output
		if ctx.Err() != nil { // this returns non-nil when context has been closed via cancellation or timeout or whatever
			return
		}
		err := coreBoyar.OnConfigChange(flags.Timeout, cfg, flags.MaxReloadTimeDelay)
		if err != nil {
			logger.Error("error executing configuration", log.Error(err))
			cfgFetcher.Resend()
		}
	}))

	supervisor.Supervise(cfgFetcher.Start(ctx))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	supervisor.WaitUntilShutdown(shutdownCtx)
	return nil
}
