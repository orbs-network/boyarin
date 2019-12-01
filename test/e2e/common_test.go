package e2e

import (
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"os"
	"time"
)

const HttpPort = 8080
const GossipPort = 4400

const WaitForBlockTimeout = 3 * time.Minute

func getConfigPath() string {
	configPath := "../../e2e-config/"
	if configPathFromEnv := os.Getenv("E2E_CONFIG"); configPathFromEnv != "" {
		configPath = configPathFromEnv
	}

	return configPath
}

func getKeyPairConfigForNode(i int, addressOnly bool) []byte {
	cfg, err := config.NewKeysConfig(fmt.Sprintf("%s/node%d/keys.json", getConfigPath(), i))
	if err != nil {
		panic(err)
	}

	return cfg.JSON(addressOnly)
}

func getHttpPortForVChain(nodeIndex int, vchainId int) int {
	return HttpPort + vchainId + nodeIndex
}

func getGossipPortForVChain(nodeIndex int, vchainId int) int {
	return GossipPort + vchainId + nodeIndex
}
