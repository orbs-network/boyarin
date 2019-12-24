package services

import (
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/supervized"
	"github.com/orbs-network/scribe/log"
)

func Execute(flags *config.Flags, logger log.Logger) error {
	if flags.ConfigUrl == "" {
		return fmt.Errorf("--config-url is a required parameter for provisioning flow")
	}

	if flags.KeyPairConfigPath == "" {
		return fmt.Errorf("--keys is a required parameter for provisioning flow")
	}

	WatchAndReportServicesStatus(logger)

	cfgFetcher := NewConfigurationPollService(flags, logger)
	coreBoyar := NewCoreBoyarService(logger)

	// wire cfg and boyar
	supervized.GoForever(func(first bool) {
		cfg := <-cfgFetcher.Output
		err := coreBoyar.OnConfigChange(flags.Timeout, cfg, flags.MaxReloadTimeDelay)
		if err != nil {
			logger.Error("error executing configuration", log.Error(err))
			cfgFetcher.Resend()
		}
	})

	cfgFetcher.Start()

	// block forever
	<-make(chan interface{})

	return nil
}
