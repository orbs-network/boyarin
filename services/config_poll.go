package services

import (
	"context"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"time"
)

type ConfigurationPollService struct {
	flags        *config.Flags
	logger       log.Logger
	output       chan config.NodeConfiguration
	Output       <-chan config.NodeConfiguration
	configCache  *utils.CacheFilter
	errorHandler govnr.Errorer
}

func NewConfigurationPollService(flags *config.Flags, logger log.Logger) *ConfigurationPollService {
	output := make(chan config.NodeConfiguration)
	return &ConfigurationPollService{
		flags:        flags,
		logger:       logger,
		output:       output,
		Output:       output,
		configCache:  utils.NewCacheFilter(),
		errorHandler: utils.NewLogErrors(logger),
	}
}

func (service *ConfigurationPollService) Start(ctx context.Context) govnr.ShutdownWaiter {
	return govnr.Forever(ctx, "configuration polling service", service.errorHandler, func() {
		defer func() {
			<-time.After(service.flags.PollingInterval)
		}()
		cfg, err := config.GetConfiguration(service.flags)
		if err != nil {
			service.logger.Error("invalid configuration", log.Error(err))
			return
		}

		if service.configCache.CheckNewValue(cfg) {
			service.output <- cfg
		} else {
			service.logger.Info("configuration has not changed")
		}
	})
}

func (service *ConfigurationPollService) Resend() {
	service.configCache.Clear()
}
