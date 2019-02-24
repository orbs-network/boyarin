package ethereum

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

type DeployingEthereumConnection interface {
	EthereumConnection
	DeployEthereumContract(auth *bind.TransactOpts, abijson string, bytecode string, params ...interface{}) (*common.Address, *bind.BoundContract, error)
}

func (c *connectorCommon) DeployEthereumContract(auth *bind.TransactOpts, abijson string, bytecode string, params ...interface{}) (*common.Address, *bind.BoundContract, error) {
	client, err := c.getContractCaller()
	if err != nil {
		return nil, nil, err
	}

	// deploy
	parsedAbi, err := abi.JSON(strings.NewReader(abijson))
	if err != nil {
		return nil, nil, err
	}
	address, _, contract, err := bind.DeployContract(auth, parsedAbi, common.FromHex(bytecode), client, params...)
	if err != nil {
		return nil, nil, err
	}

	return &address, contract, nil
}
