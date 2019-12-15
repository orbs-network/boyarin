package services

import (
	"context"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/supervized"
	"github.com/orbs-network/scribe/log"
	"time"
)

const BoyarConfigHash = "boyar config"

func RunCoreBoyarService(flags *config.Flags, logger log.Logger) {
	// Even if something crashed, things still were provisioned, meaning the cache should stay
	configCache := config.NewCache()
	<-supervized.GoForever(func(first bool) {
		defer func() {
			<-time.After(flags.PollingInterval)
		}()
		cfg, err := config.GetConfiguration(flags)
		if err != nil {
			logger.Error("invalid configuration", log.Error(err))
			return
		}
		hash := cfg.Hash()
		if hash == configCache.Get(BoyarConfigHash) {
			logger.Error("configuration has not changed")
			return
		}
		// random delay when provisioning change (that is, not bootstrap flow)
		if !first {
			randomDelay(cfg, flags.MaxReloadTimeDelay, logger)
		}
		ctx, cancel := context.WithTimeout(context.Background(), flags.Timeout)
		defer cancel()

		err = boyar.Flow(ctx, cfg, configCache, logger)
		if err != nil {
			logger.Error("error during execution", log.Error(err))
		}
		configCache.Put(BoyarConfigHash, hash)
	})
}

func randomDelay(cfg config.NodeConfiguration, maxDelay time.Duration, logger log.Logger) {
	reloadTimeDelay := cfg.ReloadTimeDelay(maxDelay)
	logger.Info("waiting to apply new configuration", log.String("delay", maxDelay.String()))
	<-time.After(reloadTimeDelay)
}
