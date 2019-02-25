package e2e

import (
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestE2EWithDockerSwarm(t *testing.T) {
	skipUnlessSwarmIsEnabled(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	h := newHarness(t, swarm)

	h.startChain(t)
	defer h.stopChain(t)

	waitForBlock(t, h.getMetrics, 3, 2*time.Minute)
}

func TestE2EKeepVolumesBetweenReloadsWithSwarm(t *testing.T) {
	skipUnlessSwarmIsEnabled(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	h := newHarness(t, swarm)

	h.startChain(t)
	defer h.stopChain(t)

	waitForBlock(t, h.getMetricsForPort(8081), 10, 2*time.Minute)

	expectedBlockHeight, err := getBlockHeight(h.getMetricsForPort(8081))
	require.NoError(t, err)

	h.startChainInstance(t, 1)

	waitForBlock(t, h.getMetricsForPort(8081), expectedBlockHeight, 2*time.Minute)
}
