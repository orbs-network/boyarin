package adapter

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerSwarm "github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

const DEFAULT_DOCKER_IMAGE = "busybox"

func TestDockerSwarm_GetStatusIfUnableToStart(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)
	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)

		serviceId := startDefunctContainer(t)
		defer destroyDefunctContainer(t, serviceId)

		swarm, err := NewDockerSwarm(OrchestratorOptions{}, log.GetLogger())
		require.NoError(t, err)

		require.True(t, helpers.Eventually(30*time.Second, func() bool {
			status, err := swarm.GetStatus(context.TODO(), 5*time.Second)
			require.NoError(t, err)
			for _, s := range status {
				if s.Name == DEFUNCT_NAME {
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
		defer destroyDefunctContainer(t, serviceId)

		swarm, err := NewDockerSwarm(OrchestratorOptions{}, log.GetLogger())
		require.NoError(t, err)

		require.True(t, helpers.Eventually(30*time.Second, func() bool {
			status, err := swarm.GetStatus(context.TODO(), 5*time.Second)
			require.NoError(t, err)
			for _, s := range status {
				if s.Name == RELOADING_NAME {
					t.Log("polling reloadingBusybox:", "s.Error=", s.Error, "s.Logs=", s.Logs)
					return strings.Contains(s.Error, "non-zero exit") && strings.Contains(s.Logs, "I can not be contained")
				}
			}

			return false
		}), "should be able to retrieve logs from constantly reloading container")
	})
}

func TestDockerSwarm_GetStatusOfUnhealthyContainer(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)
	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)

		serviceId := startUnhealthyContainer(t)
		defer destroyDefunctContainer(t, serviceId)

		swarm, err := NewDockerSwarm(OrchestratorOptions{}, log.GetLogger())
		require.NoError(t, err)

		require.True(t, helpers.Eventually(30*time.Second, func() bool {
			status, err := swarm.GetStatus(context.TODO(), 5*time.Second)
			require.NoError(t, err)
			for _, s := range status {
				if s.Name == UNHEALTHY_NAME {
					t.Log("polling reloadingBusybox:", "s.Error=", s.Error, "s.Logs=", s.Logs)
					if s.Debug.ContainerState != nil && len(s.Debug.ContainerState.Health.Log) > 0 {
						return strings.Contains(s.Error, "task: non-zero exit (137): dockerexec: unhealthy container") &&
							strings.Contains(s.Debug.ContainerState.Health.Log[0].Output, "HEALTHCHECK FAILED")
					}
				}
			}

			return false
		}), "should be able to retrieve logs from constantly reloading container")
	})
}

const DEFUNCT_NAME = "DefunctContainer"

func startDefunctContainer(t *testing.T) (serviceId string) {
	client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(DOCKER_API_VERSION))
	require.NoError(t, err)

	spec := dockerSwarm.ServiceSpec{
		TaskTemplate: dockerSwarm.TaskSpec{
			ContainerSpec: &dockerSwarm.ContainerSpec{
				Image:   DEFAULT_DOCKER_IMAGE,
				Command: []string{"this-program-does-not-exist"},
			},
		},
	}
	spec.Name = DEFUNCT_NAME

	resp, err := client.ServiceCreate(context.Background(), spec, types.ServiceCreateOptions{})
	require.NoError(t, err)

	return resp.ID
}

const RELOADING_NAME = "ReloadingContainer"

func startReloadingContainer(t *testing.T) (serviceId string) {
	client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(DOCKER_API_VERSION))
	require.NoError(t, err)

	spec := dockerSwarm.ServiceSpec{
		TaskTemplate: dockerSwarm.TaskSpec{
			ContainerSpec: &dockerSwarm.ContainerSpec{
				Image:   DEFAULT_DOCKER_IMAGE,
				Command: []string{"sh", "-c", "echo I can not be contained && exit 999"},
			},
		},
	}
	spec.Name = RELOADING_NAME

	resp, err := client.ServiceCreate(context.Background(), spec, types.ServiceCreateOptions{})
	require.NoError(t, err)

	return resp.ID
}

const UNHEALTHY_NAME = "UnhealthyContainer"

func startUnhealthyContainer(t *testing.T) (serviceId string) {
	client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(DOCKER_API_VERSION))
	require.NoError(t, err)

	spec := dockerSwarm.ServiceSpec{
		TaskTemplate: dockerSwarm.TaskSpec{
			ContainerSpec: &dockerSwarm.ContainerSpec{
				Image:   DEFAULT_DOCKER_IMAGE,
				Command: []string{"sh", "-c", "sleep 10000"},
				Healthcheck: &container.HealthConfig{
					Interval: 100 * time.Millisecond,
					Retries:  1,
					Test: []string{"CMD-SHELL",
						"sh -c \"echo HEALTHCHECK FAILED; exit 1\""},
				},
			},
		},
	}
	spec.Name = UNHEALTHY_NAME

	resp, err := client.ServiceCreate(context.Background(), spec, types.ServiceCreateOptions{})
	require.NoError(t, err)

	return resp.ID
}

func destroyDefunctContainer(t *testing.T, serviceId string) {
	client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(DOCKER_API_VERSION))
	require.NoError(t, err)

	_ = client.ServiceRemove(context.Background(), serviceId)
}
