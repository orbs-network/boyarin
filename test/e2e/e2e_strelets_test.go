package e2e

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
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

func TestCreateSignerService(t *testing.T) {
	helpers.WithContext(func(ctx context.Context) {
		helpers.InitSwarmEnvironment(t, ctx)
		client, err := dockerClient.NewEnvClient()
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
