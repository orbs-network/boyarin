package services

import (
	"context"
	"encoding/json"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/scribe/log"
)

func buildBootstrapConfig(configUrl string, ethereumEndpoint string) string {
	cfg := map[string]interface{}{
		"services": map[string]interface{}{
			"config": map[string]interface{}{
				"Port": 7666,
				"DockerConfig": map[string]interface{}{
					"Image": "orbsnetwork/network-state-reader",
					"Tag":   "latest",
					"Pull":  false,
				},
				"Management": map[string]interface{}{
					"ethereumEndpoint":     ethereumEndpoint,
					"boyarLegacyBootstrap": configUrl,
				},
			},
		},
	}

	raw, _ := json.Marshal(cfg)
	return string(raw)
}

func Bootstrap(ctx context.Context, flags *config.Flags, logger log.Logger) error {
	cfg, err := config.NewStringConfigurationSource(buildBootstrapConfig(flags.ConfigUrl, flags.EthereumEndpoint), "", flags.KeyPairConfigPath)
	if err != nil {
		return err
	}

	logger.Info("bootstrapping from default config")

	coreBoyar := NewCoreBoyarService(logger)
	return coreBoyar.OnConfigChange(ctx, cfg)
}
