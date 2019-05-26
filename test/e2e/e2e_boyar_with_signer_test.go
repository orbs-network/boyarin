package e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestE2ESingleVchainWithSignerWithSwarmAndBoyar(t *testing.T) {
	//helpers.SkipUnlessSwarmIsEnabled(t)
	removeAllDockerVolumes(t)
	defer removeAllServices(t)

	for i := 1; i <= 3; i++ {
		swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
		require.NoError(t, err)

		s := strelets.NewStrelets(swarm)

		vchains := getBoyarVchains(i, 42)
		boyarConfig := getBoyarConfigWithSigner(vchains)
		cfg, err := config.NewStringConfigurationSource(string(boyarConfig), "")
		cfg.SetKeyConfigPath(fmt.Sprintf("%s/node%d/keys.json", getConfigPath(), i))
		require.NoError(t, err)

		cache := config.NewCache()
		b := boyar.NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())
		err = b.ProvisionVirtualChains(context.Background())
		require.NoError(t, err)
		err = b.ProvisionServices(context.Background())
		require.NoError(t, err)
	}

	helpers.WaitForBlock(t, helpers.GetMetricsForPort(8125), 3, WAIT_FOR_BLOCK_TIMEOUT)
}
