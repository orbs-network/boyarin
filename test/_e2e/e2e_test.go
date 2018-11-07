package e2e

import (
	"fmt"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestE2E(t *testing.T) {
	h := newHarness(t)

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
