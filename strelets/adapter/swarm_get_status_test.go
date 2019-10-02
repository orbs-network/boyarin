package adapter

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestDockerSwarm_GetStatusIfUnableToStart(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	serviceId := startDefunctBusybox(t)
	defer destroyDefunctBusybox(t, serviceId)

	swarm, err := NewDockerSwarm(OrchestratorOptions{})
	require.NoError(t, err)

	require.True(t, helpers.Eventually(15*time.Second, func() bool {
		status, err := swarm.GetStatus(context.TODO(), 30*time.Second)
		if err != nil {
			return false
		}
		for _, s := range status {
			return s.Name == "defunctBusybox" && strings.Contains(s.Error, "executable file not found")
		}

		return false
	}), "should be able to retrieve error from container that fails to start")
}

func TestDockerSwarm_GetStatusIfExitsImmediately(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)

	serviceId := startReloadingBusybox(t)
	defer destroyDefunctBusybox(t, serviceId)

	swarm, err := NewDockerSwarm(OrchestratorOptions{})
	require.NoError(t, err)

	require.True(t, helpers.Eventually(15*time.Second, func() bool {
		status, err := swarm.GetStatus(context.TODO(), 30*time.Second)
		if err != nil {
			return false
		}
		for _, s := range status {
			return s.Name == "reloadingBusybox" &&
				strings.Contains(s.Error, "non-zero exit") &&
				strings.Contains(s.Logs, "I can not be contained")
		}

		return false
	}), "should be able to retrieve logs from constantly reloading container")
}

func startDefunctBusybox(t *testing.T) (serviceId string) {
	client, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))
	require.NoError(t, err)

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   "busybox",
				Command: []string{"this-program-does-not-exist"},
			},
		},
	}
	spec.Name = "defunctBusybox"

	resp, err := client.ServiceCreate(context.Background(), spec, types.ServiceCreateOptions{})
	require.NoError(t, err)

	return resp.ID
}

func startReloadingBusybox(t *testing.T) (serviceId string) {
	client, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))
	require.NoError(t, err)

	spec := swarm.ServiceSpec{
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: &swarm.ContainerSpec{
				Image:   "busybox",
				Command: []string{"sh", "-c", "echo I can not be contained && exit 999"},
			},
		},
	}
	spec.Name = "reloadingBusybox"

	resp, err := client.ServiceCreate(context.Background(), spec, types.ServiceCreateOptions{})
	require.NoError(t, err)

	return resp.ID
}

func destroyDefunctBusybox(t *testing.T, serviceId string) error {
	client, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))
	require.NoError(t, err)

	return client.ServiceRemove(context.Background(), serviceId)
}
