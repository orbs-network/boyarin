package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func getBoyarVchains(httpPort int, gossipPort int, nodeIndex int, vchainIds ...int) []*strelets.VirtualChain {
	var chains []*strelets.VirtualChain

	for _, vchainId := range vchainIds {
		chain := &strelets.VirtualChain{
			Id:         strelets.VirtualChainId(vchainId),
			HttpPort:   httpPort + vchainId + nodeIndex,
			GossipPort: gossipPort + vchainId + nodeIndex,
			DockerConfig: strelets.DockerImageConfig{
				ContainerNamePrefix: fmt.Sprintf("node%d", nodeIndex),
				Image:               "orbs",
				Tag:                 "export",
				Pull:                false,
			},
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
			Key:  key,
			IP:   ip,
			Port: gossipPort + int(vchains[0].Id) + i + 1,
		})
	}

	configMap["network"] = network
	configMap["chains"] = vchains

	jsonConfig, _ := json.Marshal(configMap)
	return jsonConfig
}

const HTTP_PORT = 8080
const GOSSIP_PORT = 4400

func provisionVchains(t *testing.T, h *harness, s strelets.Strelets, httpPort int, gossipPort int, i int, vchainIds ...int) {
	vchains := getBoyarVchains(httpPort, gossipPort, i, vchainIds...)
	boyarConfig := getBoyarConfig(gossipPort, vchains)
	config, err := boyar.NewStringConfigurationSource(string(boyarConfig))
	require.NoError(t, err)

	cache := make(boyar.BoyarConfigCache)
	b := boyar.NewBoyar(s, config, cache, fmt.Sprintf("%s/node%d/keys.json", h.configPath, i))
	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	//err = s.UpdateReverseProxy(context.Background(), vchains, test.LocalIP())
	//require.NoError(t, err)
}

func TestE2EProvisionMultipleVchainsWithSwarmAndBoyar(t *testing.T) {
	skipUnlessSwarmIsEnabled(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	h := newHarness(t, swarm)

	s := strelets.NewStrelets(swarm)

	for i := 1; i <= 3; i++ {
		provisionVchains(t, h, s, HTTP_PORT, GOSSIP_PORT, i, 42, 92)
	}
	// FIXME boyar should take care of it, not the harness
	defer h.stopChains(t, 42, 92)
	//defer realDocker.RemoveContainer(context.Background(), "http-api-reverse-proxy")

	waitForBlock(t, h.getMetricsForPort(8125), 3, 2*time.Minute)
	waitForBlock(t, h.getMetricsForPort(8175), 0, 20*time.Second)
}

func TestE2EAddNewVirtualChainWithSwarmAndBoyar(t *testing.T) {
	skipUnlessSwarmIsEnabled(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	h := newHarness(t, swarm)

	s := strelets.NewStrelets(swarm)

	for i := 1; i <= 3; i++ {
		provisionVchains(t, h, s, HTTP_PORT, GOSSIP_PORT, i, 42)
	}
	// FIXME boyar should take care of it, not the harness
	defer h.stopChains(t, 42)
	//defer swarm.RemoveContainer(context.Background(), "http-api-reverse-proxy")

	waitForBlock(t, h.getMetricsForPort(8125), 3, 1*time.Minute)

	for i := 1; i <= 3; i++ {
		provisionVchains(t, h, s, HTTP_PORT+1000, GOSSIP_PORT+1000, i, 42, 92)
	}
	defer h.stopChains(t, 92)

	waitForBlock(t, h.getMetricsForPort(9125), 3, 2*time.Minute)
	waitForBlock(t, h.getMetricsForPort(9175), 0, 20*time.Second)
}
