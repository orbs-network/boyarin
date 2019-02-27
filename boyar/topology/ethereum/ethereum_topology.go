package ethereum

import (
	"context"
	"github.com/orbs-network/boyarin/strelets"
)

func GetEthereumTopology(ctx context.Context, ethereumEndpoint string, topologyContractAddress string) ([]*strelets.FederationNode, error) {
	connection := NewEthereumRpcConnection(&ethereumConnectorConfig{
		endpoint: ethereumEndpoint,
	})

	address, err := StringToEthereumAddress(topologyContractAddress)
	if err != nil {
		return nil, err
	}

	packedOutput, err := CallTopologyContract(ctx, connection, &address)
	if err != nil {
		return nil, err
	}

	rawTopology, err := ABIExtractTopology(packedOutput)
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
