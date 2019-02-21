package topology

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ipToString(t *testing.T) {
	require.Equal(t, "172.10.93.43", IpToString([4]byte{172, 10, 93, 43}))
}

const NODE_ADDRESS_1 = "6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9"
const NODE_ADDESSS_2 = "d27e2e7398e2582f63d0800330010b3e58952ff6"
const NODE_IP_1 = "172.10.93.43"
const NODE_IP_2 = "172.10.93.92"

func TestConvert(t *testing.T) {
	firstAddress, _ := common.NewMixedcaseAddressFromString(NODE_ADDRESS_1)
	secondAddress, _ := common.NewMixedcaseAddressFromString(NODE_ADDESSS_2)

	topology := Convert(&ethereum.RawTopology{
		IpAddresses: [][4]byte{
			{172, 10, 93, 43},
			{172, 10, 93, 92},
		},
		NodeAddresses: []common.Address{
			firstAddress.Address(),
			secondAddress.Address(),
		},
	})

	require.EqualValues(t, []*strelets.FederationNode{
		{
			IP:  NODE_IP_1,
			Key: NODE_ADDRESS_1,
		},
		{
			IP:  NODE_IP_2,
			Key: NODE_ADDESSS_2,
		},
	}, topology)
}
