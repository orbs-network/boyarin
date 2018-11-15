package e2e

import (
	"fmt"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestE2E(t *testing.T) {
	docker := NewMockDockerAdapter()
	h := newHarness(t, docker)

	for i := 1; i <= 3; i++ {
		docker.mock.On("GetContainerConfiguration", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		docker.mock.On("StoreConfiguration", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
		docker.mock.On("RunContainer", mock.Anything, fmt.Sprintf("node%d-chain-42", i), mock.Anything).Once()
		docker.mock.On("RemoveContainer", mock.Anything, fmt.Sprintf("node%d-chain-42", i)).Once()
	}

	h.startChain(t)
	h.stopChain(t)

	docker.VerifyMocks(t)
}

func TestE2EWithRealDocker(t *testing.T) {
	if os.Getenv("ENABLE_DOCKER") != "true" {
		t.Skip("skipping test, real docker disabled")
	}

	realDocker, err := adapter.NewDockerAPI()
	require.NoError(t, err)
	h := newHarness(t, realDocker)

	h.startChain(t)
	defer h.stopChain(t)

	require.True(t, test.Eventually(10*time.Second, func() bool {
		metrics, err := h.getMetrics()
		if err != nil {
			return false
		}

		blockHeight := metrics["BlockStorage.BlockHeight"].(map[string]interface{})["Value"].(float64)
		fmt.Println("blockHeight", blockHeight)

		return blockHeight == 3
	}))
}

func TestE2EWithDockerSwarm(t *testing.T) {
	if os.Getenv("ENABLE_SWARM") != "true" {
		t.Skip("skipping test, docker swarm disabled")
	}

	realDocker, err := adapter.NewDockerSwarm()
	require.NoError(t, err)
	h := newHarness(t, realDocker)

	h.startChain(t)
	defer h.stopChain(t)

	require.True(t, test.Eventually(10*time.Second, func() bool {
		metrics, err := h.getMetrics()
		if err != nil {
			return false
		}

		blockHeight := metrics["BlockStorage.BlockHeight"].(map[string]interface{})["Value"].(float64)
		fmt.Println("blockHeight", blockHeight)

		return blockHeight == 3
	}))
}
