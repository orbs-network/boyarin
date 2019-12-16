package services

import (
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/supervized"
	"github.com/orbs-network/scribe/log"
	"time"
)

type ConfigurationPollService struct {
	flags                *config.Flags
	logger               log.Logger
	output               chan config.NodeConfiguration
	Output               <-chan config.NodeConfiguration
	LastConfigCacheToken string
}

func NewConfigurationPollService(flags *config.Flags, logger log.Logger) *ConfigurationPollService {
	output := make(chan config.NodeConfiguration)
	return &ConfigurationPollService{
		flags:                flags,
		logger:               logger,
		output:               output,
		Output:               output,
		LastConfigCacheToken: "init value - cache miss",
	}
}

func (service *ConfigurationPollService) Start() {
	supervized.GoForever(func(first bool) {
		defer func() {
			<-time.After(service.flags.PollingInterval)
		}()
		cfg, err := config.GetConfiguration(service.flags)
		if err != nil {
			service.logger.Error("invalid configuration", log.Error(err))
			return
		}
		hash := cfg.Hash()
		if hash == service.LastConfigCacheToken {
			service.logger.Error("configuration has not changed")
			return
		}

		service.output <- cfg
	})
}
