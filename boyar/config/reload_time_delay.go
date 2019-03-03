package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io/ioutil"
	"math/rand"
	"time"
)

func (n *nodeConfigurationContainer) ReloadTimeDelay(maxDelay time.Duration) time.Duration {
	if maxDelay == 0 {
		return 0
	}

	data, err := ioutil.ReadFile(n.keyConfigPath)
	if err != nil {
		return maxDelay
	}

	hash := sha256.Sum256(data)
	buf := bytes.NewBuffer(hash[:])
	seed, err := binary.ReadVarint(buf)
	if err != nil {
		return maxDelay
	}

	randomDelay := rand.New(rand.NewSource(seed)).Int63n(maxDelay.Nanoseconds())
	return time.Duration(randomDelay)
}
