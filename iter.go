package evmabi

import (
	"errors"
	"iter"

	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
)

type EventType string

const (
	TupleStart EventType = "TUPLE_START"
	TupleEnd   EventType = "TUPLE_END"
	ArrayStart EventType = "ARRAY_START"
	ArrayEnd   EventType = "ARRAY_END"
	Key        EventType = "KEY"
	Value      EventType = "VALUE"
)

type Event struct {
	Type    EventType
	ABIType eth_abi.Type
	Len     int
	Index   int
	Key     string
	Value   interface{}
}

type YieldFunc func(*Event, error) bool

func Yield(fn YieldFunc, evt *Event) error {
	if !fn(evt, nil) {
		return ErrIterStop
	}

	return nil
}

func DecodeArguments(data []byte, args eth_abi.Arguments) iter.Seq2[*Event, error] {
	return func(yield func(*Event, error) bool) {
		var err = decodeArguments(data, args, yield)

		if err == nil || errors.Is(err, ErrIterStop) {
			return
		}

		yield(nil, err)
	}
}

func DecodeValue(data []byte, t eth_abi.Type) iter.Seq2[*Event, error] {
	return func(yield func(*Event, error) bool) {
		var err = decodeValue(data, t, 0, yield)

		if err == nil || errors.Is(err, ErrIterStop) {
			return
		}

		yield(nil, err)
	}
}
