package ethereum

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"sync"
)

type EthereumRpcConnection struct {
	connectorCommon

	config ethereumAdapterConfig

	mu struct {
		sync.Mutex
		client *ethclient.Client
	}
}

func NewEthereumRpcConnection(config ethereumAdapterConfig) *EthereumRpcConnection {
	rpc := &EthereumRpcConnection{
		config: config,
	}
	rpc.getContractCaller = func() (caller EthereumCaller, e error) {
		return rpc.dialIfNeededAndReturnClient()
	}
	return rpc
}

func (rpc *EthereumRpcConnection) dial() error {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	if rpc.mu.client != nil {
		return nil
	}
	if client, err := ethclient.Dial(rpc.config.EthereumEndpoint()); err != nil {
		return err
	} else {
		rpc.mu.client = client
	}
	return nil
}

func (rpc *EthereumRpcConnection) dialIfNeededAndReturnClient() (*ethclient.Client, error) {
	if rpc.mu.client == nil {
		if err := rpc.dial(); err != nil {
			return nil, err
		}
	}
	return rpc.mu.client, nil
}

func (rpc *EthereumRpcConnection) InSync(ctx context.Context) (bool, error) {
	client, err := rpc.dialIfNeededAndReturnClient()
	if err != nil {
		return false, err
	}

	progress, err := client.SyncProgress(ctx)
	if err != nil {
		return false, err
	} else if progress == nil {
		return true, nil
	}

	return false, nil
}

func StringToEthereumAddress(input string) (common.Address, error) {
	address, err := common.NewMixedcaseAddressFromString(input)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to parse topology contract ethereum address: %s", err)
	}

	return address.Address(), nil
}
