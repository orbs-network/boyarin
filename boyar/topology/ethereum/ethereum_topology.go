package ethereum

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
)

func GetEthereumTopology(ctx context.Context, ethereumEndpoint string, topologyContractAddress string) ([]*strelets.FederationNode, error) {
	connection := NewEthereumRpcConnection(&ethereumConnectorConfig{
		endpoint: ethereumEndpoint,
	})

	if ok, err := connection.InSync(ctx); err != nil {
		return nil, fmt.Errorf("failed to retrieve topology: %s", err)
	} else if !ok {
		return nil, fmt.Errorf("failed to retrieve topology: ethereum node is not synced yet")
	}

	address, err := StringToEthereumAddress(topologyContractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to parse topology contract address: %s", err)
	}

	packedOutput, err := CallTopologyContract(ctx, connection, &address)
	if err != nil {
		return nil, fmt.Errorf("failed to call topology contract: %s", err)
	}

	rawTopology, err := ABIExtractTopology(packedOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack topology contract output: %s", err)
	}

	return rawTopology.FederationNodes(), nil
}

type ethereumConnectorConfig struct {
	endpoint string
}

func (c *ethereumConnectorConfig) EthereumEndpoint() string {
	return c.endpoint
}
