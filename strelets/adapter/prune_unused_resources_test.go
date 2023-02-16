package adapter

import (
	"context"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPruneUnusedResources(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)

		logger := log.GetLogger()
		orchestrator := &dockerSwarmOrchestrator{
			client: helpers.DockerClient(t),
			options: &OrchestratorOptions{
				StorageDriver:    LOCAL_DRIVER,
				StorageMountType: "bind",
			},
			logger: logger}

		err := orchestrator.PruneUnusedResources(ctx)
		require.NoError(t, err)

	})
}
