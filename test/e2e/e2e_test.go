package e2e

import (
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestE2EWithDockerSwarm(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	h := newHarness(t, swarm)

	h.startChain(t)
	defer h.stopChain(t)

	waitForBlock(t, h.getMetrics, 3, WAIT_FOR_BLOCK_TIMEOUT)
}

func TestE2EKeepVolumesBetweenReloadsWithSwarm(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	h := newHarness(t, swarm)

	h.startChain(t)
	defer h.stopChain(t)

	waitForBlock(t, h.getMetricsForPort(8081), 10, WAIT_FOR_BLOCK_TIMEOUT)

	expectedBlockHeight, err := getBlockHeight(h.getMetricsForPort(8081))
	require.NoError(t, err)

	h.startChainInstance(t, 1)

	waitForBlock(t, h.getMetricsForPort(8081), expectedBlockHeight, WAIT_FOR_BLOCK_TIMEOUT)
}
