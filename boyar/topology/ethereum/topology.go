package ethereum

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/orbs-network/boyarin/strelets"
	"net"
	"strings"
)

type RawTopology struct {
	NodeAddresses []common.Address
	IpAddresses   [][4]byte
}

const TopologyContractABI = `
[
  {
    "constant": false,
    "inputs": [],
    "name": "getNetworkTopology",
    "outputs": [
      {
        "name": "NodeAddresses",
        "type": "address[]"
      },
      {
        "name": "IpAddresses",
        "type": "bytes4[]"
      }
    ],
    "payable": false,
    "stateMutability": "nonpayable",
    "type": "function"
  }
]`

const TopologyContractMethodName = "getNetworkTopology"

func IpToString(ip [4]byte) string {
	return net.IPv4(ip[0], ip[1], ip[2], ip[3]).String()
}

func EthereumToOrbsAddress(eth string) string {
	return strings.ToLower(eth[2:])
}

func (rawTopology *RawTopology) FederationNodes() (federationNodes []*strelets.FederationNode) {
	for index, address := range rawTopology.NodeAddresses {
		federationNodes = append(federationNodes, &strelets.FederationNode{
			Address: EthereumToOrbsAddress(address.Hex()),
			IP:      IpToString(rawTopology.IpAddresses[index]),
		})
	}

	return
}

func ABIExtractTopology(packedOutput []byte) (*RawTopology, error) {
	parsedABI, err := abi.JSON(strings.NewReader(TopologyContractABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %s", err)
	}

	value := new(RawTopology)
	if err := parsedABI.Unpack(value, TopologyContractMethodName, packedOutput); err != nil {
		return nil, fmt.Errorf("failed to unpack output: %s", err)
	}

	return value, nil
}

func CallTopologyContract(ctx context.Context, connection DeployingEthereumConnection, contractAddress *common.Address) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(TopologyContractABI))
	if err != nil {
		return nil, err
	}

	ethCallData, err := parsedABI.Pack(TopologyContractMethodName)
	if err != nil {
		return nil, err
	}

	return connection.CallContract(ctx, contractAddress.Bytes(), ethCallData, nil)
}
