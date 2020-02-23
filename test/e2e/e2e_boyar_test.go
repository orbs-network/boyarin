package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/services"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

func getBoyarVchains(nodeIndex int, vchainIds ...int) []*config.VirtualChain {
	var chains []*config.VirtualChain

	for _, vchainId := range vchainIds {
		chain := &config.VirtualChain{
			Id:         config.VirtualChainId(vchainId),
			HttpPort:   getHttpPortForVChain(nodeIndex, vchainId),
			GossipPort: getGossipPortForVChain(nodeIndex, vchainId),
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

func getConfigMap(vchains []*config.VirtualChain) map[string]interface{} {
	ip := helpers.LocalIP()

	configMap := make(map[string]interface{})
	var network []*strelets.FederationNode
	for i, key := range helpers.NodeAddresses() {
		network = append(network, &strelets.FederationNode{
			Address: key,
			IP:      ip,
			Port:    GossipPort + int(vchains[0].Id) + i + 1,
		})
	}

	configMap["network"] = network
	configMap["chains"] = vchains

	return configMap
}

func getBoyarConfig(vchains []*config.VirtualChain) []byte {
	configMap := getConfigMap(vchains)
	jsonConfig, _ := json.Marshal(configMap)
	return jsonConfig
}

func getBoyarConfigWithSigner(i int, vchains []*config.VirtualChain) []byte {
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

// Tests boyar.Flow as close as it gets to production starting up
func TestE2EWithFullFlowAndDisabledSimilarVchainId(t *testing.T) {
	helpers.SkipOnCI(t)
	logger := helpers.DefaultTestLogger()
	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)

		for i := 1; i <= 4; i++ {
			vchains := getBoyarVchains(i, 1000, 92, 100)
			vchains[len(vchains)-1].Disabled = true // Check for namespace clashes: 100 will be removed but 1000 should be intact

			boyarConfig := getBoyarConfig(vchains)
			logger.Info(fmt.Sprintf("node %d config %s", i, string(boyarConfig)))
			cfg, err := config.NewStringConfigurationSource(string(boyarConfig), helpers.LocalEthEndpoint()) // ethereum endpoint is optional
			require.NoError(t, err)
			cfg.SetKeyConfigPath(fmt.Sprintf("%s/node%d/keys.json", getConfigPath(), i))

			err = services.NewCoreBoyarService(logger).OnConfigChange(ctx, cfg)
			require.NoError(t, err)
		}

		helpers.WaitForBlock(t, helpers.GetMetricsForPort(getHttpPortForVChain(1, 1000)), 3, WaitForBlockTimeout)
		helpers.WaitForBlock(t, helpers.GetMetricsForPort(getHttpPortForVChain(1, 92)), 0, WaitForBlockTimeout)

		_, err := helpers.GetMetricsForPort(getHttpPortForVChain(1, 100))() // port for vcid 100
		require.Error(t, err)
		require.Regexp(t, ".*connection refused.*", err.Error())
	})
}
