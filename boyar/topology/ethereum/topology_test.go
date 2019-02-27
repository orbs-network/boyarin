package ethereum

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_IpToString(t *testing.T) {
	require.Equal(t, "172.10.93.43", IpToString([4]byte{172, 10, 93, 43}))
}

const NODE_ADDRESS_1 = "6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9"
const NODE_ADDESSS_2 = "d27e2e7398e2582f63d0800330010b3e58952ff6"
const NODE_IP_1 = "172.10.93.43"
const NODE_IP_2 = "172.10.93.92"

func Test_RawTopology_FederationNodes(t *testing.T) {
	firstAddress, _ := common.NewMixedcaseAddressFromString(NODE_ADDRESS_1)
	secondAddress, _ := common.NewMixedcaseAddressFromString(NODE_ADDESSS_2)

	federationNodes := (&RawTopology{
		IpAddresses: [][4]byte{
			{172, 10, 93, 43},
			{172, 10, 93, 92},
		},
		NodeAddresses: []common.Address{
			firstAddress.Address(),
			secondAddress.Address(),
		},
	}).FederationNodes()

	require.EqualValues(t, []*strelets.FederationNode{
		{
			IP:      NODE_IP_1,
			Address: NODE_ADDRESS_1,
		},
		{
			IP:      NODE_IP_2,
			Address: NODE_ADDESSS_2,
		},
	}, federationNodes)
}

func Test_ABIExtractTopology(t *testing.T) {
	VALID_PACKED_OUTPUT := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 128, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 255, 255, 255, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	rawTopology, err := ABIExtractTopology(VALID_PACKED_OUTPUT)

	require.NoError(t, err)
	require.EqualValues(t, &RawTopology{
		IpAddresses: [][4]byte{
			{255, 255, 255, 255},
		},
		NodeAddresses: []common.Address{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}, rawTopology)
}

func Test_ABIExtractTopologyWithErrors(t *testing.T) {
	INVALID_PACKED_OUTPUT := []byte{255, 0, 0, 0, 123, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 128, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 255, 255, 255, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	rawTopology, err := ABIExtractTopology(INVALID_PACKED_OUTPUT)

	require.EqualError(t, err, "failed to unpack output: abi: cannot marshal in to go slice: offset 115339776401686340910430194812341638118309611963264502320445993504662469214304 would go over slice boundary (len=192)")
	require.Nil(t, rawTopology)
}
