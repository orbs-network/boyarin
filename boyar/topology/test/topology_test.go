package test

import (
	"context"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTopologyIntegrationWithGanache(t *testing.T) {
	skipUnlessEthereumIsEnabled(t)

	test.WithContext(func(ctx context.Context) {
		h := newRpcEthereumConnectorHarness(t, getConfig())

		contractAddress, err := h.deployContract(ethereum.TopologyContractABI, TopologyContractBytecode)
		require.NoError(t, err, "failed deploying topology to Ethereum")
		require.NotNil(t, contractAddress, "contract address is empty")

		topology, err := ethereum.GetEthereumTopology(ctx, getConfig().EthereumEndpoint(), contractAddress.Hex())
		require.NoError(t, err, "failed to retrieve topology")

		t.Log(topology[0])

		require.EqualValues(t, "255.255.255.255", topology[0].IP, "should match expected IP")
		require.EqualValues(t, "0000000000000000000000000000000000000000", topology[0].Address, "should match expected public address")
	})
}

func TestTopologyIntegrationWithGanacheUnhappyFlow(t *testing.T) {
	skipUnlessEthereumIsEnabled(t)

	test.WithContext(func(ctx context.Context) {
		badContractAddress := "0000000000000000000000000000000000000000"

		topology, err := ethereum.GetEthereumTopology(ctx, "malformed ethereum endpoint", badContractAddress)
		require.EqualError(t, err, "failed to retrieve topology: dial unix malformed ethereum endpoint: connect: no such file or directory")
		require.Nil(t, topology, "topology should be empty")

		topology, err = ethereum.GetEthereumTopology(ctx, getConfig().EthereumEndpoint(), badContractAddress)
		require.EqualError(t, err, "failed to call topology contract: no contract code at given address")
		require.Nil(t, topology, "topology should be empty")

	})
}
