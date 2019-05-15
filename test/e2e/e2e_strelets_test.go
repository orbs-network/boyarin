package e2e

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
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
		HttpPort:     HTTP_PORT + i,
		GossipPort:   GOSSIP_PORT + i,
		DockerConfig: dockerConfig(fmt.Sprintf("node%d", i)),
		Config:       helpers.ChainConfigWithGenesisValidatorAddresses(),
	}
}

func stopChainWithStrelets(t *testing.T, s strelets.Strelets, i int) {
	err := s.RemoveVirtualChain(context.Background(), &strelets.RemoveVirtualChainInput{
		VirtualChain: chain(i),
	})

	require.NoError(t, err)
	fmt.Println(fmt.Sprintf("stopped node%d", i))
}

func startChainWithStrelets(t *testing.T, s strelets.Strelets, i int) {
	localIP := helpers.LocalIP()
	ctx := context.Background()

	err := s.ProvisionVirtualChain(ctx, &strelets.ProvisionVirtualChainInput{
		NodeAddress:       strelets.NodeAddress(helpers.NodeAddresses()[i-1]),
		VirtualChain:      chain(i),
		Peers:             peers(localIP),
		KeyPairConfigPath: fmt.Sprintf("%s/node%d/keys.json", getConfigPath(), i),
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
				Memory: 256,
				CPUs:   0.25,
			},
			Reservations: strelets.Resource{
				Memory: 128,
				CPUs:   0.25,
			},
		},
	}
}

func peers(ip string) *strelets.PeersMap {
	peers := make(strelets.PeersMap)

	for i, key := range helpers.NodeAddresses() {
		peers[strelets.NodeAddress(key)] = &strelets.Peer{
			IP:   ip,
			Port: 4400 + i + 1,
		}
	}

	return &peers
}

func TestE2EWithDockerSwarm(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)
	removeAllDockerVolumes(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	s := strelets.NewStrelets(swarm)

	for i := 1; i <= 3; i++ {
		startChainWithStrelets(t, s, i)
		defer stopChainWithStrelets(t, s, i)
	}

	helpers.WaitForBlock(t, helpers.GetMetricsForPort(8081), 3, WAIT_FOR_BLOCK_TIMEOUT)
}

func TestE2EKeepVolumesBetweenReloadsWithSwarm(t *testing.T) {
	helpers.SkipUnlessSwarmIsEnabled(t)
	removeAllDockerVolumes(t)

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	s := strelets.NewStrelets(swarm)

	for i := 1; i <= 3; i++ {
		startChainWithStrelets(t, s, i)
		defer stopChainWithStrelets(t, s, i)
	}

	helpers.WaitForBlock(t, helpers.GetMetricsForPort(8081), 10, WAIT_FOR_BLOCK_TIMEOUT)

	expectedBlockHeight, err := helpers.GetBlockHeight(helpers.GetMetricsForPort(8081))
	require.NoError(t, err)

	stopChainWithStrelets(t, s, 1)
	time.Sleep(3 * time.Second)
	startChainWithStrelets(t, s, 1)

	helpers.WaitForBlock(t, helpers.GetMetricsForPort(8081), expectedBlockHeight, WAIT_FOR_BLOCK_TIMEOUT)
}

func TestCreateServiceSysctls(t *testing.T) {
	t.Skip("not supported on Mac or CI, relies on 19.03 beta features")
	helpers.SkipUnlessSwarmIsEnabled(t)

	client, err := client.NewClientWithOpts(client.WithVersion(adapter.DOCKER_API_VERSION))
	if err != nil {
		t.Errorf("could not connect to docker: %s", err)
		t.FailNow()
	}
	defer client.Close()

	ctx := context.Background()

	swarm, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	require.NoError(t, err)
	s := strelets.NewStrelets(swarm)

	startChainWithStrelets(t, s, 1)
	defer stopChainWithStrelets(t, s, 1)

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
	tasks, err := client.TaskList(ctx, types.TaskListOptions{
		Filters: filter,
	})
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	// verify that the container has the sysctl option set
	ctnr, err := client.ContainerInspect(ctx, tasks[0].Status.ContainerStatus.ContainerID)
	require.NoError(t, err)
	require.EqualValuesf(t, adapter.GetSysctls(), ctnr.HostConfig.Sysctls, "failed to set container sysctls")

	// verify that the task has the sysctl option set in the task object
	require.EqualValuesf(t, adapter.GetSysctls(), tasks[0].Spec.ContainerSpec.Sysctls, "failed to set container spec sysctls")

	// verify that the service also has the sysctl set in the spec.
	service, _, err := client.ServiceInspectWithRaw(ctx, "node1-chain-42-stack", types.ServiceInspectOptions{})
	require.NoError(t, err)
	require.EqualValuesf(t,
		adapter.GetSysctls(), service.Spec.TaskTemplate.ContainerSpec.Sysctls,
		"failed to set service sysctls",
	)
}
