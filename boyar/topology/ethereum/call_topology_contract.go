package ethereum

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

func CallTopologyContract(ctx context.Context, connection DeployingEthereumConnection, contractAddress *common.Address) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(TopologyContractABI))
	if err != nil {
		return nil, err
	}

	ethCallData, err := ABIPackFunctionInputArguments(parsedABI, TopologyContractMethodName, nil)
	if err != nil {
		return nil, err
	}

	return connection.CallContract(ctx, contractAddress.Bytes(), ethCallData, nil)
}
