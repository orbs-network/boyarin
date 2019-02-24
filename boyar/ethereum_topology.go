package boyar

import (
	"context"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
	"github.com/orbs-network/boyarin/strelets"
)

func GetEthereumTopology(ctx context.Context, ethereumEndpoint string, topologyContractAddress string) ([]*strelets.FederationNode, error) {
	connection := ethereum.NewEthereumRpcConnection(&ethereumConnectorConfig{
		endpoint: ethereumEndpoint,
	})

	address, err := ethereum.StringToEthereumAddress(topologyContractAddress)
	if err != nil {
		return nil, err
	}

	packedOutput, err := ethereum.CallTopologyContract(ctx, connection, &address)
	if err != nil {
		return nil, err
	}

	rawTopology, err := ethereum.ABIExtractTopology(packedOutput)
	if err != nil {
		return nil, err
	}

	return rawTopology.FederationNodes(), nil
}

type ethereumConnectorConfig struct {
	endpoint string
}

func (c *ethereumConnectorConfig) EthereumEndpoint() string {
	return c.endpoint
}
