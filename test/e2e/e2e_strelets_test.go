package e2e

import (
	"context"
	"fmt"
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
