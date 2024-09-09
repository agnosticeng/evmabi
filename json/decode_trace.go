package json

import (
	"fmt"

	"github.com/bytedance/sonic/ast"
	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
)

func DecodeTrace(input []byte, output []byte, method eth_abi.Method) (ast.Node, error) {
	if len(method.Outputs) == 0 && len(output) > 0 {
		return ast.Node{}, fmt.Errorf("trace has output data but method has no outputs")
	}

	inputs, err := DecodeArguments(input[4:], method.Inputs)

	if err != nil {
		return ast.Node{}, err
	}

	outputs, err := DecodeArguments(output, method.Outputs)

	if err != nil {
		return ast.Node{}, err
	}

	return ast.NewObject([]ast.Pair{
		ast.NewPair("signature", ast.NewString(method.Sig)),
		ast.NewPair("inputs", inputs),
		ast.NewPair("outputs", outputs),
	}), nil
}
