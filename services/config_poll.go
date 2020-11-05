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
	errorHandler govnr.Errorer
}

// FIXME completely remove in the future
func NewConfigurationPollService(flags *config.Flags, logger log.Logger) *ConfigurationPollService {
	output := make(chan config.NodeConfiguration)
	return &ConfigurationPollService{
		flags:        flags,
		logger:       logger,
		output:       output,
		Output:       output,
		errorHandler: utils.NewLogErrors("configuration polling", logger),
	}
}

func (service *ConfigurationPollService) Start(ctx context.Context) govnr.ShutdownWaiter {
	handle := govnr.Forever(ctx, "configuration polling service", service.errorHandler, func() {
		defer func() {
			select {
			case <-ctx.Done():
			case <-time.After(service.flags.PollingInterval):
			}
		}()
		cfg, err := config.GetConfiguration(service.flags)
		if err != nil {
			service.logger.Error("invalid configuration", log.Error(err))
			return
		}

		service.output <- cfg
	})
	go func() {
		<-handle.Done()
		close(service.output)
	}()
	return handle
}
