package ethereum

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
)

func ABIPackFunctionInputArguments(abi abi.ABI, functionName string, args []interface{}) ([]byte, error) {
	return abi.Pack(functionName, args...)
}

func ABIUnpackFunctionOutputArguments(abi abi.ABI, out interface{}, functionName string, packedOutput []byte) error {
	return abi.Unpack(out, functionName, packedOutput)
}
