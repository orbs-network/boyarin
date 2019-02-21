package test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestRawTopologyE2EWithGanache(t *testing.T) {
	//if !runningWithDocker() {
	//	t.Skip("this test relies on external components - ganache, and will be skipped unless running in docker")
	//}

	test.WithContext(func(ctx context.Context) {
		h := newRpcEthereumConnectorHarness(t, getConfig())

		contractAddress, err := h.deployContract(ethereum.TopologyContractABI, ethereum.TopologyContractBytecode)
		require.NoError(t, err, "failed deploying topology to Ethereum")
		require.NotNil(t, contractAddress, "contract address is empty")

		fmt.Println(hexutil.Encode(contractAddress[:]))

		packedOutput, err := ethereum.CallTopologyContract(ctx, h.rpcAdapter, contractAddress)
		require.NoError(t, err, "expecting call to succeed")
		require.True(t, len(packedOutput) > 0, "expecting packedOutput to have some data")

		value, err := ethereum.ABIExtractTopology(packedOutput)
		require.NoError(t, err, "could not unpack results")

		require.Len(t, value.NodeAddresses, 1)
		require.Len(t, value.IpAddresses, 1)

		ip := ethereum.IpToString(value.IpAddresses[0])
		nodeAddress := ethereum.EthereumToOrbsAddress(value.NodeAddresses[0].Hex())

		fmt.Println(ip)
		fmt.Println(nodeAddress)

		require.EqualValues(t, "255.255.255.255", ip)
		require.EqualValues(t, "0000000000000000000000000000000000000000", nodeAddress)
	})
}

func TestTopologyE2EWithGanache(t *testing.T) {
	//if !runningWithDocker() {
	//	t.Skip("this test relies on external components - ganache, and will be skipped unless running in docker")
	//}

	test.WithContext(func(ctx context.Context) {
		h := newRpcEthereumConnectorHarness(t, getConfig())

		contractAddress, err := h.deployContract(ethereum.TopologyContractABI, ethereum.TopologyContractBytecode)
		require.NoError(t, err, "failed deploying topology to Ethereum")
		require.NotNil(t, contractAddress, "contract address is empty")

		topology, err := boyar.GetEthereumTopology(ctx, getConfig().EthereumEndpoint(), contractAddress.Hex())
		require.NoError(t, err, "failed to retrieve topology")

		fmt.Println(topology[0])

		require.EqualValues(t, "255.255.255.255", topology[0].IP)
		require.EqualValues(t, "0000000000000000000000000000000000000000", topology[0].Key)
	})
}

func runningWithDocker() bool {
	return os.Getenv("EXTERNAL_TEST") == "true"
}

func getConfig() *ethereumConnectorConfig {
	var cfg ethereumConnectorConfig

	return &ethereumConnectorConfig{
		endpoint:      "http://localhost:7545",
		privateKeyHex: "7a16631b19e5a7d121f13c3ece279c10c996ff14d8bebe609bf1eca41211b291", // mnemonic for this pk: pet talent sugar must audit chief biology trash change wheat educate bone
	}

	if endpoint := os.Getenv("ETHEREUM_ENDPOINT"); endpoint != "" {
		cfg.endpoint = endpoint
	}

	if privateKey := os.Getenv("ETHEREUM_PRIVATE_KEY"); privateKey != "" {
		cfg.privateKeyHex = privateKey
	}

	return &cfg
}
