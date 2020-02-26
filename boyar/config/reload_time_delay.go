package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"time"
)

func (c *nodeConfigurationContainer) ReloadTimeDelay(maxDelay time.Duration) time.Duration {
	overrideDelay := c.value.OrchestratorOptions.MaxReloadTimedDelay()
	if c.value.OrchestratorOptions.MaxReloadTimedDelay() != 0 {
		return overrideDelay
	}

	if maxDelay == 0 {
		return 0
	}

	cfg, err := c.readKeysConfig()
	if err != nil {
		return maxDelay
	}

	hash := sha256.Sum256([]byte(cfg.PrivateKey()))
	buf := bytes.NewBuffer(hash[:])
	seed, err := binary.ReadVarint(buf)
	if err != nil {
		return maxDelay
	}

	randomDelay := rand.New(rand.NewSource(seed)).Int63n(maxDelay.Nanoseconds())
	return time.Duration(randomDelay)
}
