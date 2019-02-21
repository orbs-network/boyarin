package test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	//"go/types"
	"os"
	"strings"
	"testing"
)

func TestTopologyE2EWithGanache(t *testing.T) {
	// the idea of this test is to make sure that the entire 'call-from-ethereum' logic works on a spedific timestamp and different states in time (blocks)
	// it requires ganache or some other simulation to transact

	//if !runningWithDocker() {
	//	t.Skip("this test relies on external components - ganache, and will be skipped unless running in docker")
	//}
	contractABI := `
  [
    {
      "constant": false,
      "inputs": [],
      "name": "getNetworkTopology",
      "outputs": [
        {
          "name": "nodeAddresses",
          "type": "address[]"
        },
        {
          "name": "ipAddresses",
          "type": "bytes4[]"
        }
      ],
      "payable": false,
      "stateMutability": "nonpayable",
      "type": "function"
    }
  ]`

	contractBytecode := `0x6080604052602060405190810160405280600073ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681525060009060016100519291906100e6565b5060206040519081016040528063ffffffff7c0100000000000000000000000000000000000000000000000000000000027bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19167bffffffffffffffffffffffffffffffffffffffffffffffffffffffff191681525060019060016100d3929190610170565b503480156100e057600080fd5b506102b0565b82805482825590600052602060002090810192821561015f579160200282015b8281111561015e5782518260006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555091602001919060010190610106565b5b50905061016c919061023a565b5090565b828054828255906000526020600020906007016008900481019282156102295791602002820160005b838211156101f757835183826101000a81548163ffffffff02191690837c0100000000000000000000000000000000000000000000000000000000900402179055509260200192600401602081600301049283019260010302610199565b80156102275782816101000a81549063ffffffff02191690556004016020816003010492830192600103026101f7565b505b509050610236919061027d565b5090565b61027a91905b8082111561027657600081816101000a81549073ffffffffffffffffffffffffffffffffffffffff021916905550600101610240565b5090565b90565b6102ad91905b808211156102a957600081816101000a81549063ffffffff021916905550600101610283565b5090565b90565b610267806102bf6000396000f3fe608060405234801561001057600080fd5b5060043610610048576000357c010000000000000000000000000000000000000000000000000000000090048063204296731461004d575b600080fd5b6100556100f4565b604051808060200180602001838103835285818151815260200191508051906020019060200280838360005b8381101561009c578082015181840152602081019050610081565b50505050905001838103825284818151815260200191508051906020019060200280838360005b838110156100de5780820151818401526020810190506100c3565b5050505090500194505050505060405180910390f35b606080600060018180548060200260200160405190810160405280929190818152602001828054801561017c57602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019060010190808311610132575b505050505091508080548060200260200160405190810160405280929190818152602001828054801561022c57602002820191906000526020600020906000905b82829054906101000a90047c0100000000000000000000000000000000000000000000000000000000027bffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916815260200190600401906020826003010492830192600103820291508084116101bd5790505b5050505050905091509150909156fea165627a7a7230582003a21a54c68895828bd1e91558253a15b56454346e17a3121cf2ea20bece83890029`

	test.WithContext(func(ctx context.Context) {
		h := newRpcEthereumConnectorHarness(t, getConfig())

		contractAddress, err := h.deployContract(contractABI, contractBytecode)
		require.NoError(t, err, "failed deploying topology to Ethereum")
		require.NotNil(t, contractAddress, "contract address is empty")

		fmt.Println(hexutil.Encode(contractAddress[:]))

		methodToCall := "getNetworkTopology"

		parsedABI, err := abi.JSON(strings.NewReader(string(contractABI)))
		require.NoError(t, err, "abi parse failed for simple storage contract")

		ethCallData, err := ethereum.ABIPackFunctionInputArguments(parsedABI, methodToCall, nil)
		require.NoError(t, err, "this means we couldn't pack the params for ethereum, something is broken with the harness")

		packedOutput, err := h.rpcAdapter.CallContract(ctx, contractAddress.Bytes(), ethCallData, nil)
		require.NoError(t, err, "expecting call to succeed")
		require.True(t, len(packedOutput) > 0, "expecting packedOutput to have some data")

		value := new(struct { // this is the expected return type of that ethereum call for the SimpleStorage contract getValues
			NodeAddresses []byte
			IpAddresses   []byte
		})

		//var nodeAddresses [][]byte
		//var ipAddresses [][]byte
		//value := types.NewTuple(
		//	types.NewVar(0, nil, "nodeAddresses", nil),
		//	types.NewVar(0, nil, "ipAddresses", nil))
		//var value types.Tuple
		err = ethereum.ABIUnpackFunctionOutputArguments(parsedABI, &value, methodToCall, packedOutput)
		require.NoError(t, err, "could not unpack results")

		fmt.Println(value)

	})
}

func runningWithDocker() bool {
	return os.Getenv("EXTERNAL_TEST") == "true"
}

func getConfig() *ethereumConnectorConfigForTests {
	var cfg ethereumConnectorConfigForTests

	return &ethereumConnectorConfigForTests{
		endpoint:      "http://localhost:7545",
		privateKeyHex: "9b7632951d1889b3a1252d42235b6cdcaa8755fc0b1bcc40108132a573c43366",
	}

	if endpoint := os.Getenv("ETHEREUM_ENDPOINT"); endpoint != "" {
		cfg.endpoint = endpoint
	}

	if privateKey := os.Getenv("ETHEREUM_PRIVATE_KEY"); privateKey != "" {
		cfg.privateKeyHex = privateKey
	}

	return &cfg
}
