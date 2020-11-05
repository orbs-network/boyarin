package services

import (
	"context"
	"errors"
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

	cfg, err := config.NewStringConfigurationSource(string(data), flags.EthereumEndpoint, flags.KeyPairConfigPath, flags.WithNamespace)
	if err != nil {
		return nil, err
	}

	// FIXME find a better way to pass the flags
	newFlags := &config.Flags{
		ConfigUrl: cfg.OrchestratorOptions().DynamicManagementConfig.Url,

		KeyPairConfigPath: flags.KeyPairConfigPath,

		Timeout:         flags.Timeout,
		PollingInterval: flags.PollingInterval,

		SSLPrivateKeyPath:  flags.SSLPrivateKeyPath,
		SSLCertificatePath: flags.SSLCertificatePath,

		EthereumEndpoint: flags.EthereumEndpoint,

		OrchestratorOptions: flags.OrchestratorOptions,

		LogFilePath:     flags.LogFilePath,
		StatusFilePath:  flags.StatusFilePath,
		MetricsFilePath: flags.MetricsFilePath,

		WithNamespace: flags.WithNamespace,

		AutoUpdate:          flags.AutoUpdate,
		ShutdownAfterUpdate: flags.AutoUpdate,
		BoyarBinaryPath:     flags.BoyarBinaryPath,
	}

	coreBoyar := NewCoreBoyarService(logger)
	if shouldExit := coreBoyar.CheckForUpdates(flags, cfg.OrchestratorOptions().ExecutableImage); shouldExit {
		logger.Info("shutting down after updating boyar binary")
		return flags, errors.New("restart needed after an update")
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	WatchAndReportStatusAndMetrics(ctxWithCancel, logger, flags)

	return newFlags, coreBoyar.OnConfigChange(ctx, cfg)
}
