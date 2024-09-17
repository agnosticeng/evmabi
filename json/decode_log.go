package json

import (
	"fmt"

	"github.com/bytedance/sonic/ast"
	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
)

func DecodeLog(topics [][]byte, input []byte, event eth_abi.Event) (ast.Node, error) {
	var indexed, unindexed = SplitInputs(event.Inputs)

	// mismatch btw num of indexed fields and num of topics
	if len(indexed) != (len(topics) - 1) {
		return ast.Node{}, fmt.Errorf("event has %d indexed inputs but log has %d topics", len(indexed), (len(topics) - 1))
	}

	// log has data but abi field does not have unindexed fields
	if len(unindexed) > 0 && len(input) == 0 {
		return ast.Node{}, fmt.Errorf("event have unindexed inputs but log has no data")
	}

	inputs, err := DecodeArguments(input, unindexed)

	if err != nil {
		return ast.Node{}, fmt.Errorf("failed to decode non-indexed fields: %w", err)
	}

	length, err := inputs.Len()

	if err != nil {
		return ast.Node{}, err
	}

	if len(unindexed) != length {
		return ast.Node{}, fmt.Errorf("wrong number of unindexed args")
	}

	for i, input := range indexed {
		v, err := DecodeValue(topics[i+1], input.Type)

		if err != nil {
			return ast.Node{}, err
		}

		inputs.Set(input.Name, v)
	}

	return ast.NewObject([]ast.Pair{
		ast.NewPair("signature", ast.NewString(event.Sig)),
		ast.NewPair("inputs", inputs),
	}), nil
}

func SplitInputs(inputs []eth_abi.Argument) ([]eth_abi.Argument, []eth_abi.Argument) {
	var (
		indexed   []eth_abi.Argument
		unindexed []eth_abi.Argument
	)

	for _, input := range inputs {
		if input.Indexed {
			indexed = append(indexed, input)
		} else {
			unindexed = append(unindexed, input)
		}
	}

	return indexed, unindexed
}
