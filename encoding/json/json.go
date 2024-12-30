package json

import (
	"errors"
	"fmt"
	"iter"

	"github.com/agnosticeng/evmabi/encoding"
	"github.com/bytedance/sonic/ast"
	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"
)

var ErrEndOfSeq = errors.New("end of seq")

func DecodeArguments(data []byte, args eth_abi.Arguments) (ast.Node, error) {
	var (
		it         = encoding.DecodeArguments(data, args)
		next, stop = iter.Pull2[*encoding.Event, error](it)
	)

	defer stop()
	return ReadValue(next)
}

func DecodeValue(data []byte, t eth_abi.Type) (ast.Node, error) {
	var (
		it         = encoding.DecodeValue(data, t)
		next, stop = iter.Pull2[*encoding.Event, error](it)
	)

	defer stop()
	return ReadValue(next)
}

func ReadTuple(next func() (*encoding.Event, error, bool), length int) (ast.Node, error) {
	var pairs []ast.Pair

	for i := 0; i < length; i++ {
		k, err := ReadKey(next)

		if err != nil {
			return ast.Node{}, err
		}

		v, err := ReadValue(next)

		if err != nil {
			return ast.Node{}, err
		}

		pairs = append(pairs, ast.NewPair(k, v))
	}

	evt, err := pullEvent(next)

	if err != nil {
		return ast.Node{}, err
	}

	if evt.Type != encoding.TupleEnd {
		return ast.Node{}, fmt.Errorf("wrong event type; wanted ArrayEnd but got %s", evt.Type)
	}

	return ast.NewObject(pairs), nil

}

func ReadArray(next func() (*encoding.Event, error, bool), length int) (ast.Node, error) {
	var nodes []ast.Node

	for i := 0; i < length; i++ {
		v, err := ReadValue(next)

		if err != nil {
			return ast.Node{}, err
		}

		nodes = append(nodes, v)
	}

	evt, err := pullEvent(next)

	if err != nil {
		return ast.Node{}, err
	}

	if evt.Type != encoding.ArrayEnd {
		return ast.Node{}, fmt.Errorf("wrong event type; wanted ArrayEnd but got %s", evt.Type)
	}

	return ast.NewArray(nodes), nil
}

func ReadKey(next func() (*encoding.Event, error, bool)) (string, error) {
	var evt, err = pullEvent(next)

	if err != nil {
		return "", err
	}

	if evt.Type != encoding.Key {
		return "", fmt.Errorf("wrong event type; wanted Key but got %s", evt.Type)
	}

	return evt.Key, nil
}

func ReadValue(next func() (*encoding.Event, error, bool)) (ast.Node, error) {
	var evt, err = pullEvent(next)

	if err != nil {
		return ast.Node{}, err
	}

	switch evt.Type {
	case encoding.Value:
		switch evt.ABIType.T {
		case eth_abi.BytesTy, eth_abi.FixedBytesTy:
			return ast.NewAny(hexutil.Bytes(evt.Value.([]byte))), nil

		case eth_abi.IntTy:
			var i = evt.Value.(*uint256.Int)

			if i.Sign() == -1 {
				return ast.NewAny(fmt.Sprintf("-%d", uint256.NewInt(0).Neg(i))), nil
			}

			return ast.NewAny(evt.Value), nil

		default:
			return ast.NewAny(evt.Value), nil
		}
	case encoding.TupleStart:
		return ReadTuple(next, evt.Len)
	case encoding.ArrayStart:
		return ReadArray(next, evt.Len)
	default:
		return ast.Node{}, fmt.Errorf("wrong event type; wanted Value|TupleStart|ArrayStart but got %s", evt.Type)
	}
}

func pullEvent(next func() (*encoding.Event, error, bool)) (*encoding.Event, error) {
	var evt, err, ok = next()

	if !ok {
		return nil, ErrEndOfSeq
	}

	if err != nil {
		return nil, err
	}

	return evt, nil
}
