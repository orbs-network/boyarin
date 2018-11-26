package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestE2E(t *testing.T) {
	docker := NewMockDockerAdapter()
	h := newHarness(t, docker)

	for i := 1; i <= 3; i++ {
		docker.mock.On("Prepare", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		docker.runner.mock.On("Run", mock.Anything).Once()
		docker.mock.On("RemoveContainer", mock.Anything, fmt.Sprintf("node%d-chain-42", i)).Once()
	}

	h.startChain(t)
	h.stopChain(t)

	docker.VerifyMocks(t)
}

func TestE2EWithDocker(t *testing.T) {
	skipUnlessDockerIsEnabled(t)

	realDocker, err := adapter.NewDockerAPI("_tmp")
	require.NoError(t, err)
	h := newHarness(t, realDocker)

	h.startChain(t)
	defer h.stopChain(t)

	waitForBlock(t, h.getMetrics, 3, 15*time.Second)
}

func TestE2EWithDockerAndBoyar(t *testing.T) {
	skipUnlessDockerIsEnabled(t)

	realDocker, err := adapter.NewDockerAPI("_tmp")
	require.NoError(t, err)
	h := newHarness(t, realDocker)

	s := strelets.NewStrelets("_tmp", realDocker)

	configMap := make(map[string]interface{})

	ip := test.LocalIP()

	var network []*strelets.FederationNode
	for i, key := range test.PublicKeys() {
		network = append(network, &strelets.FederationNode{
			Key:  key,
			IP:   ip,
			Port: 4400 + i + 1,
		})
	}

	configMap["network"] = network

	for i := 1; i <= 3; i++ {
		chain := &strelets.VirtualChain{
			Id:         42,
			HttpPort:   8080 + i,
			GossipPort: 4400 + i,
			DockerConfig: &strelets.DockerImageConfig{
				ContainerNamePrefix: fmt.Sprintf("node%d", i),
				Image:               "orbs",
				Tag:                 "export",
				Pull:                false,
			},
		}

		configMap["chains"] = []*strelets.VirtualChain{chain}

		jsonConfig, _ := json.Marshal(configMap)
		config, err := boyar.NewStringConfigurationSource(string(jsonConfig))
		require.NoError(t, err)

		b := boyar.NewBoyar(s, config, fmt.Sprintf("%s/node%d/keys.json", h.configPath, i))
		err = b.ProvisionVirtualChains(context.Background())
		require.NoError(t, err)
	}
	// FIXME boyar should take care of it, not the harness
	defer h.stopChain(t)

	waitForBlock(t, h.getMetrics, 3, 20*time.Second)
}

func TestE2EWithDockerSwarm(t *testing.T) {
	skipUnlessSwarmIsEnabled(t)

	realDocker, err := adapter.NewDockerSwarm()
	require.NoError(t, err)
	h := newHarness(t, realDocker)

	h.startChain(t)
	defer h.stopChain(t)

	waitForBlock(t, h.getMetrics, 3, 20*time.Second)
}
