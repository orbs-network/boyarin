package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func getBoyarConfig(httpPort int, gossipPort int, nodeIndex int, vchainIds ...int) []byte {
	ip := test.LocalIP()

	configMap := make(map[string]interface{})
	var network []*strelets.FederationNode
	for i, key := range test.PublicKeys() {
		network = append(network, &strelets.FederationNode{
			Key:  key,
			IP:   ip,
			Port: gossipPort + i + 1,
		})
	}

	configMap["network"] = network

	var chains []*strelets.VirtualChain

	for _, vchainId := range vchainIds {
		chain := &strelets.VirtualChain{
			Id:         strelets.VirtualChainId(vchainId),
			HttpPort:   httpPort + nodeIndex,
			GossipPort: gossipPort + nodeIndex,
			DockerConfig: &strelets.DockerImageConfig{
				ContainerNamePrefix: fmt.Sprintf("node%d", nodeIndex),
				Image:               "orbs",
				Tag:                 "export",
				Pull:                false,
			},
		}

		chains = append(chains, chain)
	}

	configMap["chains"] = chains

	jsonConfig, _ := json.Marshal(configMap)
	return jsonConfig
}

func TestE2EWithDockerAndBoyar(t *testing.T) {
	skipUnlessDockerIsEnabled(t)

	realDocker, err := adapter.NewDockerAPI("_tmp")
	require.NoError(t, err)
	h := newHarness(t, realDocker)

	s := strelets.NewStrelets("_tmp", realDocker)

	for i := 1; i <= 3; i++ {
		boyarConfig := getBoyarConfig(8080, 4400, i, 42)
		config, err := boyar.NewStringConfigurationSource(string(boyarConfig))
		require.NoError(t, err)

		b := boyar.NewBoyar(s, config, fmt.Sprintf("%s/node%d/keys.json", h.configPath, i))
		err = b.ProvisionVirtualChains(context.Background())
		require.NoError(t, err)
	}
	// FIXME boyar should take care of it, not the harness
	defer h.stopChain(t)

	waitForBlock(t, h.getMetricsForPort(8083), 3, 20*time.Second)
}
