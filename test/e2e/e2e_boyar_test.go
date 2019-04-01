package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

func getBoyarVchains(httpPort int, gossipPort int, nodeIndex int, vchainIds ...int) []*strelets.VirtualChain {
	var chains []*strelets.VirtualChain

	for _, vchainId := range vchainIds {
		chain := &strelets.VirtualChain{
			Id:         strelets.VirtualChainId(vchainId),
			HttpPort:   httpPort + vchainId + nodeIndex,
			GossipPort: gossipPort + vchainId + nodeIndex,
			DockerConfig: strelets.DockerConfig{
				ContainerNamePrefix: fmt.Sprintf("node%d", nodeIndex),
				Image:               "orbs",
				Tag:                 "export",
				Pull:                false,
			},
			Config: helpers.ChainConfigWithGenesisValidatorAddresses(),
		}

		chains = append(chains, chain)
	}

	return chains
}

func getBoyarConfig(gossipPort int, vchains []*strelets.VirtualChain) []byte {
	ip := helpers.LocalIP()

	configMap := make(map[string]interface{})
	var network []*strelets.FederationNode
	for i, key := range helpers.NodeAddresses() {
		network = append(network, &strelets.FederationNode{
			Address: key,
			IP:      ip,
			Port:    gossipPort + int(vchains[0].Id) + i + 1,
		})
	}

	configMap["network"] = network
	configMap["chains"] = vchains

	jsonConfig, _ := json.Marshal(configMap)
	return jsonConfig
}

const HTTP_PORT = 8080
const GOSSIP_PORT = 4400

func provisionVchains(t *testing.T, h *harness, s strelets.Strelets, httpPort int, gossipPort int, i int, vchainIds ...int) (boyar.Boyar, []*strelets.VirtualChain) {
	vchains := getBoyarVchains(httpPort, gossipPort, i, vchainIds...)
	boyarConfig := getBoyarConfig(gossipPort, vchains)
	cfg, err := config.NewStringConfigurationSource(string(boyarConfig), "")
	cfg.SetKeyConfigPath(fmt.Sprintf("%s/node%d/keys.json", h.configPath, i))
	require.NoError(t, err)

	cache := make(config.BoyarConfigCache)
	b := boyar.NewBoyar(s, cfg, cache)
	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)

	return b, cfg.Chains()
}

func disableChains(t *testing.T, b boyar.Boyar, chains []*strelets.VirtualChain) {
	for _, chain := range chains {
		chain.Disabled = true
	}
	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
}

func TestE2EProvisionMultipleVchainsWithSwarmAndBoyar(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	h := newHarness(t, swarm)

	s := strelets.NewStrelets(swarm)

	for i := 1; i <= 3; i++ {
		b, chains := provisionVchains(t, h, s, HTTP_PORT, GOSSIP_PORT, i, 42, 92)
		defer disableChains(t, b, chains)
	}

	waitForBlock(t, h.getMetricsForPort(8125), 3, WAIT_FOR_BLOCK_TIMEOUT)
	waitForBlock(t, h.getMetricsForPort(8175), 0, WAIT_FOR_BLOCK_TIMEOUT)
}

func TestE2EAddNewVirtualChainWithSwarmAndBoyar(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	h := newHarness(t, swarm)

	s := strelets.NewStrelets(swarm)

	for i := 1; i <= 3; i++ {
		provisionVchains(t, h, s, HTTP_PORT, GOSSIP_PORT, i, 42)
	}

	waitForBlock(t, h.getMetricsForPort(8125), 3, WAIT_FOR_BLOCK_TIMEOUT)

	for i := 1; i <= 3; i++ {
		b, chains := provisionVchains(t, h, s, HTTP_PORT+1000, GOSSIP_PORT+1000, i, 42, 92)
		defer disableChains(t, b, chains)
	}

	waitForBlock(t, h.getMetricsForPort(9125), 3, WAIT_FOR_BLOCK_TIMEOUT)
	waitForBlock(t, h.getMetricsForPort(9175), 0, WAIT_FOR_BLOCK_TIMEOUT)
}
