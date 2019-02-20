package test

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
	"testing"
)

type harness struct {
	rpcAdapter ethereum.DeployingEthereumConnection
	address    string
	config     *ethereumConnectorConfigForTests
}

type ethereumConnectorConfigForTests struct {
	endpoint      string
	privateKeyHex string
}

func (c *ethereumConnectorConfigForTests) EthereumEndpoint() string {
	return c.endpoint
}

func (h *harness) getAddress() string {
	return h.address
}

func (h *harness) deployContract(abi string, bytecode string) (*common.Address, error) {
	auth, err := h.authFromConfig()
	if err != nil {
		return nil, err
	}
	address, _, err := h.rpcAdapter.DeployEthereumContract(auth, abi, bytecode)
	if err != nil {
		return nil, err
	}

	return address, nil
}

func newRpcEthereumConnectorHarness(tb testing.TB, cfg *ethereumConnectorConfigForTests) *harness {
	a := ethereum.NewEthereumRpcConnection(cfg)

	return &harness{
		config:     cfg,
		rpcAdapter: a,
	}
}

func (h *harness) authFromConfig() (*bind.TransactOpts, error) {
	key, err := crypto.HexToECDSA(h.config.privateKeyHex)
	if err != nil {
		return nil, err
	}

	return bind.NewKeyedTransactor(key), nil
}
