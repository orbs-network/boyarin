package e2e

import (
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
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

func TestE2EWithDockerSwarm(t *testing.T) {
	skipUnlessSwarmIsEnabled(t)

	realDocker, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	h := newHarness(t, realDocker)

	h.startChain(t)
	defer h.stopChain(t)

	waitForBlock(t, h.getMetrics, 3, 20*time.Second)
}

func TestE2EKeepVolumesBetweenReloadsWithDocker(t *testing.T) {
	skipUnlessDockerIsEnabled(t)

	realDocker, err := adapter.NewDockerAPI("_tmp")
	require.NoError(t, err)
	h := newHarness(t, realDocker)

	h.startChain(t)
	defer h.stopChain(t)

	waitForBlock(t, h.getMetricsForPort(8081), 10, 20*time.Second)

	expectedBlockHeight, err := getBlockHeight(h.getMetricsForPort(8081))
	require.NoError(t, err)

	h.startChainInstance(t, 1)

	waitForBlock(t, h.getMetricsForPort(8081), expectedBlockHeight, 3*time.Second)
}

func TestE2EKeepVolumesBetweenReloadsWithSwarm(t *testing.T) {
	skipUnlessSwarmIsEnabled(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	h := newHarness(t, swarm)

	h.startChain(t)
	defer h.stopChain(t)

	waitForBlock(t, h.getMetricsForPort(8081), 10, 25*time.Second)

	expectedBlockHeight, err := getBlockHeight(h.getMetricsForPort(8081))
	require.NoError(t, err)

	h.startChainInstance(t, 1)

	waitForBlock(t, h.getMetricsForPort(8081), expectedBlockHeight, 10*time.Second)
}
