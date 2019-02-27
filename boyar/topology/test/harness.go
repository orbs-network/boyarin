package test

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
	"os"
	"testing"
)

const TopologyContractBytecode = `0x6080604052602060405190810160405280600073ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681525060009060016100519291906100e6565b5060206040519081016040528063ffffffff7c0100000000000000000000000000000000000000000000000000000000027bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19167bffffffffffffffffffffffffffffffffffffffffffffffffffffffff191681525060019060016100d3929190610170565b503480156100e057600080fd5b506102b0565b82805482825590600052602060002090810192821561015f579160200282015b8281111561015e5782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555091602001919060010190610106565b5b50905061016c919061023a565b5090565b828054828255906000526020600020906007016008900481019282156102295791602002820160005b838211156101f757835183826101000a81548163ffffffff02191690837c0100000000000000000000000000000000000000000000000000000000900402179055509260200192600401602081600301049283019260010302610199565b80156102275782816101000a81549063ffffffff02191690556004016020816003010492830192600103026101f7565b505b509050610236919061027d565b5090565b61027a91905b8082111561027657600081816101000a81549073ffffffffffffffffffffffffffffffffffffffff021916905550600101610240565b5090565b90565b6102ad91905b808211156102a957600081816101000a81549063ffffffff021916905550600101610283565b5090565b90565b610267806102bf6000396000f3fe608060405234801561001057600080fd5b5060043610610048576000357c010000000000000000000000000000000000000000000000000000000090048063204296731461004d575b600080fd5b6100556100f4565b604051808060200180602001838103835285818151815260200191508051906020019060200280838360005b8381101561009c578082015181840152602081019050610081565b50505050905001838103825284818151815260200191508051906020019060200280838360005b838110156100de5780820151818401526020810190506100c3565b5050505090500194505050505060405180910390f35b606080600060018180548060200260200160405190810160405280929190818152602001828054801561017c57602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019060010190808311610132575b505050505091508080548060200260200160405190810160405280929190818152602001828054801561022c57602002820191906000526020600020906000905b82829054906101000a90047c0100000000000000000000000000000000000000000000000000000000027bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190600401906020826003010492830192600103820291508084116101bd5790505b5050505050905091509150909156fea165627a7a723058205aaee76c7d8da6bf5054315a2f76e49d3c4df8902a1c77bb7f86368a84fdbace0029`

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