package test

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
	"os"
	"testing"
)

type harness struct {
	rpcAdapter ethereum.DeployingEthereumConnection
	address    string
	config     *ethereumConnectorConfig
}

type ethereumConnectorConfig struct {
	endpoint      string
	privateKeyHex string
}

func (c *ethereumConnectorConfig) EthereumEndpoint() string {
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

func newRpcEthereumConnectorHarness(tb testing.TB, cfg *ethereumConnectorConfig) *harness {
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

func skipUnlessEthereumIsEnabled(t *testing.T) {
	if os.Getenv("ENABLE_ETHEREUM") != "true" {
		t.Skip("skipping test, ethereum is disabled")
	}
}

func getConfig() *ethereumConnectorConfig {
	var cfg ethereumConnectorConfig

	//return &ethereumConnectorConfig{
	//	endpoint:      "http://localhost:7545",
	//	privateKeyHex: "7a16631b19e5a7d121f13c3ece279c10c996ff14d8bebe609bf1eca41211b291", // mnemonic for this pk: pet talent sugar must audit chief biology trash change wheat educate bone
	//}

	if endpoint := os.Getenv("ETHEREUM_ENDPOINT"); endpoint != "" {
		cfg.endpoint = endpoint
	}

	if privateKey := os.Getenv("ETHEREUM_PRIVATE_KEY"); privateKey != "" {
		cfg.privateKeyHex = privateKey
	}

	return &cfg
}
