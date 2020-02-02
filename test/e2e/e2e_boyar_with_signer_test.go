package e2e

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestE2ESingleVchainWithSignerWithSwarmAndBoyar(t *testing.T) {
	// helpers.SkipOnCI(t)
	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)
		swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{}, log.GetLogger())
		require.NoError(t, err)
		s := strelets.NewStrelets(swarm)

		for i := 1; i <= 4; i++ {

			vchains := getBoyarVchains(i, 42)
			boyarConfig := getBoyarConfigWithSigner(i, vchains)
			cfg, err := config.NewStringConfigurationSource(string(boyarConfig), "")
			require.NoError(t, err)
			cfg.SetKeyConfigPath(fmt.Sprintf("%s/node%d/keys.json", getConfigPath(), i))

			b := boyar.NewBoyar(s, cfg, boyar.NewCache(), helpers.DefaultTestLogger())
			err = b.ProvisionServices(context.Background())
			require.NoError(t, err)
			err = b.ProvisionVirtualChains(context.Background())
			require.NoError(t, err)
		}

		helpers.WaitForBlock(t, helpers.GetMetricsForPort(getHttpPortForVChain(1, 42)), 3, WaitForBlockTimeout)
	})
}
