package e2e

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	swarmTypes "github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func chain(i int) *strelets.VirtualChain {
	return &strelets.VirtualChain{
		Id:           42,
		HttpPort:     HttpPort + i,
		GossipPort:   GossipPort + i,
		DockerConfig: dockerConfig(fmt.Sprintf("node%d", i)),
		Config:       helpers.ChainConfigWithGenesisValidatorAddresses(),
	}
}

func startChainWithStrelets(t *testing.T, s strelets.Strelets, i int) {
	localIP := helpers.LocalIP()
	ctx := context.Background()

	err := s.ProvisionVirtualChain(ctx, &strelets.ProvisionVirtualChainInput{
		NodeAddress:   strelets.NodeAddress(helpers.NodeAddresses()[i-1]),
		VirtualChain:  chain(i),
		Peers:         peers(localIP),
		KeyPairConfig: getKeyPairConfigForNode(i, false),
	})

	require.NoError(t, err)
	fmt.Println(fmt.Sprintf("started node%d", i))
}

func dockerConfig(node string) strelets.DockerConfig {
	return strelets.DockerConfig{
		Image:               "orbs",
		Tag:                 "export",
		Pull:                false,
		ContainerNamePrefix: node,
		Resources: strelets.DockerResources{
			Limits: strelets.Resource{
				Memory: 2048,
				CPUs:   1,
			},
			Reservations: strelets.Resource{
				Memory: 10,
				CPUs:   0.01,
			},
		},
	}
}

func peers(ip string) *strelets.PeersMap {
	peers := make(strelets.PeersMap)

	for i, key := range helpers.NodeAddresses() {
		peers[strelets.NodeAddress(key)] = &strelets.Peer{
			IP:   ip,
			Port: GossipPort + i + 1,
		}
	}

	return &peers
}

func TestE2EWithDockerSwarm(t *testing.T) {
	helpers.SkipOnCI(t)

	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)
		swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
		require.NoError(t, err)
		s := strelets.NewStrelets(swarm)

		for i := 1; i <= 4; i++ {
			startChainWithStrelets(t, s, i)
		}

		helpers.WaitForBlock(t, helpers.GetMetricsForPort(8081), 3, WaitForBlockTimeout)
	})
}

func TestE2EKeepVolumesBetweenReloadsWithSwarm(t *testing.T) {
	helpers.SkipOnCI(t)

	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)
		swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
		require.NoError(t, err)
		s := strelets.NewStrelets(swarm)

		for i := 1; i <= 4; i++ {
			startChainWithStrelets(t, s, i)
		}

		helpers.WaitForBlock(t, helpers.GetMetricsForPort(8081), 10, WaitForBlockTimeout)

		expectedBlockHeight, err := helpers.GetBlockHeight(helpers.GetMetricsForPort(8081))
		require.NoError(t, err)

		err = s.RemoveVirtualChain(context.Background(), &strelets.RemoveVirtualChainInput{
			VirtualChain: chain(1),
		})
		require.NoError(t, err)

		time.Sleep(3 * time.Second)
		startChainWithStrelets(t, s, 1)

		helpers.WaitForBlock(t, helpers.GetMetricsForPort(8081), expectedBlockHeight, WaitForBlockTimeout)
	})
}

func TestCreateServiceSysctls(t *testing.T) {
	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)

		client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(adapter.DOCKER_API_VERSION))
		if err != nil {
			t.Errorf("could not connect to docker: %s", err)
			t.FailNow()
		}
		defer client.Close()

		swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
		require.NoError(t, err)
		s := strelets.NewStrelets(swarm)

		startChainWithStrelets(t, s, 1)

		time.Sleep(5 * time.Second)

		// Straight from Docker integration test:
		// integration/service/create_test.go
		// https://github.com/moby/moby/pull/37701/files#diff-204a9536b52c895f8a02e75d2e00dd16

		// we're going to check 3 things:
		//
		//   1. Does the container, when inspected, have the sysctl option set?
		//   2. Does the task have the sysctl in the spec?
		//   3. Does the service have the sysctl in the spec?
		//
		// if all 3 of these things are true, we know that the sysctl has been
		// plumbed correctly through the engine.

		// get all of the tasks of the service, so we can get the container
		filter := filters.NewArgs()
		filter.Add("service", "node1-chain-42-stack")
		helpers.RequireEventually(t, 1*time.Minute, func(t helpers.TestingT) {
			var tasks []swarmTypes.Task
			for len(tasks) == 0 || tasks[0].Status.ContainerStatus == nil {
				tasks, err = client.TaskList(ctx, types.TaskListOptions{
					Filters: filter,
				})
				require.NoError(t, err)
				require.Len(t, tasks, 1)
			}
			// verify that the container has the sysctl option set
			ctnr, err := client.ContainerInspect(ctx, tasks[0].Status.ContainerStatus.ContainerID)
			require.NoError(t, err)
			require.EqualValuesf(t, adapter.GetSysctls(), ctnr.HostConfig.Sysctls, "failed to set container sysctls")

			// verify that the task has the sysctl option set in the task object
			require.EqualValuesf(t, adapter.GetSysctls(), tasks[0].Spec.ContainerSpec.Sysctls, "failed to set container spec sysctls")
		})

		// verify that the service also has the sysctl set in the spec.
		service, _, err := client.ServiceInspectWithRaw(ctx, "node1-chain-42-stack", types.ServiceInspectOptions{})
		require.NoError(t, err)
		require.EqualValuesf(t,
			adapter.GetSysctls(), service.Spec.TaskTemplate.ContainerSpec.Sysctls,
			"failed to set service sysctls",
		)
	})
}

func TestCreateSignerService(t *testing.T) {
	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)
		client, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(adapter.DOCKER_API_VERSION))
		if err != nil {
			t.Errorf("could not connect to docker: %s", err)
			t.FailNow()
		}
		defer client.Close()

		swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
		require.NoError(t, err)
		s := strelets.NewStrelets(swarm)

		err = s.UpdateService(ctx, &strelets.UpdateServiceInput{
			Service: &strelets.Service{
				DockerConfig: strelets.DockerConfig{
					Image:               "orbs",
					Tag:                 "signer",
					ContainerNamePrefix: "node1",
				},
			},
			KeyPairConfig: getKeyPairConfigForNode(1, false),
		})
		require.NoError(t, err)

		require.True(t, helpers.Eventually(10*time.Second, func() bool {
			filter := filters.NewArgs()
			filter.Add("service", "node1-signer-service-stack")
			tasks, err := client.TaskList(ctx, types.TaskListOptions{
				Filters: filter,
			})

			if err != nil || len(tasks) == 0 {
				return false
			}

			return tasks[0].Status.State == "running"
		}))
	})
}
