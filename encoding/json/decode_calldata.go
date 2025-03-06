package json

import (
	"fmt"

	"github.com/bytedance/sonic/ast"
	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
)

func DecodeCallData(data []byte, method eth_abi.Method) (ast.Node, error) {
	if len(data) < 4 {
		return ast.Node{}, fmt.Errorf("call data is smaller than 4 bytes")
	}

	inputs, err := DecodeArguments(data[4:], method.Inputs)

	if err != nil {
		return ast.Node{}, err
	}

	return ast.NewObject([]ast.Pair{
		ast.NewPair("signature", ast.NewString(method.Sig)),
		ast.NewPair("inputs", inputs),
	}), nil
}
