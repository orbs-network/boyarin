package services

import (
	"context"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/scribe/log"
	"io/ioutil"
)

func Bootstrap(ctx context.Context, flags *config.Flags, logger log.Logger) (*config.Flags, error) {
	logger.Info("bootstrapping from management config", log.String("config", flags.ManagementConfig))

	data, err := ioutil.ReadFile(flags.ManagementConfig)
	if err != nil {
		return nil, err
	}

	cfg, err := config.NewStringConfigurationSource(string(data), flags.EthereumEndpoint, flags.KeyPairConfigPath)
	if err != nil {
		return nil, err
	}

	flags = &config.Flags{
		ConfigUrl: cfg.OrchestratorOptions().DynamicManagementConfig.Url,

		KeyPairConfigPath: flags.KeyPairConfigPath,

		Timeout:         flags.Timeout,
		PollingInterval: flags.PollingInterval,

		SSLPrivateKeyPath:  flags.SSLPrivateKeyPath,
		SSLCertificatePath: flags.SSLCertificatePath,

		EthereumEndpoint: flags.EthereumEndpoint,

		OrchestratorOptions: flags.OrchestratorOptions,

		LogFilePath: flags.LogFilePath,
	}

	coreBoyar := NewCoreBoyarService(logger)
	return flags, coreBoyar.OnConfigChange(ctx, cfg)
}
