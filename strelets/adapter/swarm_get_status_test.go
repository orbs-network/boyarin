package adapter

import (
	"context"
	"github.com/docker/docker/api/types"
	dockerSwarm "github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestDockerSwarm_GetStatusIfUnableToStart(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)
	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)

		serviceId := startDefunctContainer(t)
		defer destroyDefunctBusybox(t, serviceId)

		swarm, err := NewDockerSwarm(OrchestratorOptions{}, log.GetLogger())
		require.NoError(t, err)

		require.True(t, helpers.Eventually(30*time.Second, func() bool {
			status, err := swarm.GetStatus(context.TODO(), 5*time.Second)
			require.NoError(t, err)
			for _, s := range status {
				if s.Name == defunctName {
					t.Log("polling reloadingBusybox:", "s.Error=", s.Error, "s.Logs=", s.Logs)
					return strings.Contains(s.Error, "executable file not found")
				}
			}

			return false
		}), "should be able to retrieve error from container that fails to start")
	})
}

func TestDockerSwarm_GetStatusIfExitsImmediately(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)
	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)

		serviceId := startReloadingContainer(t)
		defer destroyDefunctBusybox(t, serviceId)

		swarm, err := NewDockerSwarm(OrchestratorOptions{}, log.GetLogger())
		require.NoError(t, err)

		require.True(t, helpers.Eventually(30*time.Second, func() bool {
			status, err := swarm.GetStatus(context.TODO(), 5*time.Second)
			require.NoError(t, err)
			for _, s := range status {
				if s.Name == reloadingName {
					t.Log("polling reloadingBusybox:", "s.Error=", s.Error, "s.Logs=", s.Logs)
					return strings.Contains(s.Error, "non-zero exit") && strings.Contains(s.Logs, "I can not be contained")
				}
			}

			return false
		}), "should be able to retrieve logs from constantly reloading container")
	})
}

const defunctName = "DefunctContainer"

func startDefunctContainer(t *testing.T) (serviceId string) {
	client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(DOCKER_API_VERSION))
	require.NoError(t, err)

	spec := dockerSwarm.ServiceSpec{
		TaskTemplate: dockerSwarm.TaskSpec{
			ContainerSpec: &dockerSwarm.ContainerSpec{
				Image:   "alpine",
				Command: []string{"this-program-does-not-exist"},
			},
		},
	}
	spec.Name = defunctName

	resp, err := client.ServiceCreate(context.Background(), spec, types.ServiceCreateOptions{})
	require.NoError(t, err)

	return resp.ID
}

const reloadingName = "ReloadingContainer"

func startReloadingContainer(t *testing.T) (serviceId string) {
	client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(DOCKER_API_VERSION))
	require.NoError(t, err)

	spec := dockerSwarm.ServiceSpec{
		TaskTemplate: dockerSwarm.TaskSpec{
			ContainerSpec: &dockerSwarm.ContainerSpec{
				Image:   "alpine",
				Command: []string{"sh", "-c", "echo I can not be contained && exit 999"},
			},
		},
	}
	spec.Name = reloadingName

	resp, err := client.ServiceCreate(context.Background(), spec, types.ServiceCreateOptions{})
	require.NoError(t, err)

	return resp.ID
}

func destroyDefunctBusybox(t *testing.T, serviceId string) {
	client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(DOCKER_API_VERSION))
	require.NoError(t, err)

	_ = client.ServiceRemove(context.Background(), serviceId)
}
