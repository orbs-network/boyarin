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

func getBoyarVchains(nodeIndex int, vchainIds ...int) []*strelets.VirtualChain {
	var chains []*strelets.VirtualChain

	for _, vchainId := range vchainIds {
		chain := &strelets.VirtualChain{
			Id:         strelets.VirtualChainId(vchainId),
			HttpPort:   HTTP_PORT + vchainId + nodeIndex,
			GossipPort: GOSSIP_PORT + vchainId + nodeIndex,
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

func getConfigMap(vchains []*strelets.VirtualChain) map[string]interface{} {
	ip := helpers.LocalIP()

	configMap := make(map[string]interface{})
	var network []*strelets.FederationNode
	for i, key := range helpers.NodeAddresses() {
		network = append(network, &strelets.FederationNode{
			Address: key,
			IP:      ip,
			Port:    GOSSIP_PORT + int(vchains[0].Id) + i + 1,
		})
	}

	configMap["network"] = network
	configMap["chains"] = vchains

	return configMap
}

func getBoyarConfig(vchains []*strelets.VirtualChain) []byte {
	configMap := getConfigMap(vchains)
	jsonConfig, _ := json.Marshal(configMap)
	return jsonConfig
}

func getBoyarConfigWithSigner(i int, vchains []*strelets.VirtualChain) []byte {
	configMap := getConfigMap(vchains)
	configMap["services"] = strelets.Services{
		Signer: &strelets.Service{
			Port: 7777,
			DockerConfig: strelets.DockerConfig{
				ContainerNamePrefix: fmt.Sprintf("node%d", i),
				Image:               "orbs",
				Tag:                 "signer",
			},
		},
	}

	jsonConfig, _ := json.Marshal(configMap)
	return jsonConfig
}

func provisionVchains(t *testing.T, s strelets.Strelets, i int, vchainIds ...int) (boyar.Boyar, []*strelets.VirtualChain) {
	vchains := getBoyarVchains(i, vchainIds...)
	boyarConfig := getBoyarConfig(vchains)
	cfg, err := config.NewStringConfigurationSource(string(boyarConfig), "")
	cfg.SetKeyConfigPath(fmt.Sprintf("%s/node%d/keys.json", getConfigPath(), i))
	require.NoError(t, err)

	cache := config.NewCache()
	b := boyar.NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())
	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)

	return b, cfg.Chains()
}

func TestE2EProvisionMultipleVchainsWithSwarmAndBoyar(t *testing.T) {
	withCleanContext(t, func(t *testing.T) {
		swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
		require.NoError(t, err)

		s := strelets.NewStrelets(swarm)

		for i := 1; i <= 3; i++ {
			provisionVchains(t, s, i, 42, 92)
		}

		helpers.WaitForBlock(t, helpers.GetMetricsForPort(8125), 3, WAIT_FOR_BLOCK_TIMEOUT)
		helpers.WaitForBlock(t, helpers.GetMetricsForPort(8175), 0, WAIT_FOR_BLOCK_TIMEOUT)
	})
}

func TestE2EAddNewVirtualChainWithSwarmAndBoyar(t *testing.T) {
	withCleanContext(t, func(t *testing.T) {

		swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
		require.NoError(t, err)

		s := strelets.NewStrelets(swarm)

		for i := 1; i <= 3; i++ {
			provisionVchains(t, s, i, 42)
		}

		helpers.WaitForBlock(t, helpers.GetMetricsForPort(8125), 3, WAIT_FOR_BLOCK_TIMEOUT)

		for i := 1; i <= 3; i++ {
			provisionVchains(t, s, i, 42, 92)
		}

		helpers.WaitForBlock(t, helpers.GetMetricsForPort(8125), 3, WAIT_FOR_BLOCK_TIMEOUT)
		helpers.WaitForBlock(t, helpers.GetMetricsForPort(8175), 0, WAIT_FOR_BLOCK_TIMEOUT)
	})
}

// Tests boyar.Flow as close as it gets to production starting up
func TestE2EWithFullFlowAndDisabledSimilarVchainId(t *testing.T) {
	withCleanContext(t, func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			vchains := getBoyarVchains(i, 1000, 92, 100)
			vchains[len(vchains) - 1].Disabled = true // Check for namespace clashes: 100 will be removed but 1000 should be intact

			boyarConfig := getBoyarConfig(vchains)
			cfg, err := config.NewStringConfigurationSource(string(boyarConfig), "")
			cfg.SetKeyConfigPath(fmt.Sprintf("%s/node%d/keys.json", getConfigPath(), i))
			require.NoError(t, err)

			logger := helpers.DefaultTestLogger()
			cache := config.NewCache()
			err = boyar.Flow(context.Background(), cfg, cache, logger)
			require.NoError(t, err)
		}

		helpers.WaitForBlock(t, helpers.GetMetricsForPort(9081), 3, WAIT_FOR_BLOCK_TIMEOUT)
		helpers.WaitForBlock(t, helpers.GetMetricsForPort(8175), 0, WAIT_FOR_BLOCK_TIMEOUT)

		helpers.Eventually(WAIT_FOR_BLOCK_TIMEOUT, func() bool {
			_, err := helpers.GetMetricsForPort(8181)() // port for vcid 100
			return err != nil
		})
	})
}
