package adapter

import (
	"context"
	"github.com/docker/docker/api/types/mount"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func createFilesPerMount(t *testing.T, mounts []mount.Mount) {
	require.NotZero(t, len(mounts), "number of mounts should be more than 0")

	for _, m := range mounts {
		volumePath := m.Source
		pathWithDir := path.Join(volumePath, "some-dir")
		if err := os.MkdirAll(pathWithDir, 0755); err != nil {
			require.NoError(t, err)
		}

		if err := ioutil.WriteFile(path.Join(pathWithDir, "file-in-dir"), []byte("file-in-dir"), 0755); err != nil {
			require.NoError(t, err)
		}

		if err := ioutil.WriteFile(path.Join(volumePath, "regular-file"), []byte("regular-file"), 0755); err != nil {
			require.NoError(t, err)
		}
	}
}

func mountsToPaths(mounts []mount.Mount) (results []string) {
	for _, m := range mounts {
		results = append(results, m.Source)
	}

	return
}

func TestPurgeServiceData(t *testing.T) {
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

		containerName := "diamond-dogs"

		mounts, err := orchestrator.provisionServiceVolumes(ctx, containerName, nil)
		require.NoError(t, err)

		require.False(t, helpers.VerifyFilesExist(t, mountsToPaths(mounts)...))

		createFilesPerMount(t, mounts)
		require.True(t, helpers.VerifyFilesExist(t, mountsToPaths(mounts)...))

		err = orchestrator.PurgeServiceData(ctx, containerName)
		require.NoError(t, err)
		require.False(t, helpers.VerifyFilesExist(t, mountsToPaths(mounts)...))
	})
}
