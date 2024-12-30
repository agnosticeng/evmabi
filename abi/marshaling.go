package abi

import (
	"fmt"
	"strconv"

	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
)

type ArgumentMarshaling struct {
	Name         string                `json:"name,omitempty"`
	Type         string                `json:"type,omitempty"`
	InternalType string                `json:"internalType,omitempty"`
	Components   []*ArgumentMarshaling `json:"components,omitempty"`
	Indexed      bool                  `json:"indexed,omitempty"`
}

type FieldMarshaling struct {
	Type    string                `json:"type,omitempty"`
	Name    string                `json:"name,omitempty"`
	Inputs  []*ArgumentMarshaling `json:"inputs,omitempty"`
	Outputs []*ArgumentMarshaling `json:"outputs,omitempty"`

	// Status indicator which can be: "pure", "view",
	// "nonpayable" or "payable".
	StateMutability string `json:"stateMutability,omitempty"`

	// Deprecated Status indicators, but removed in v0.6.0.
	Constant bool `json:"constant,omitempty"` // True if function is either pure or view
	Payable  bool `json:"payable,omitempty"`  // True if function is payable

	// Event relevant indicator represents the event is
	// declared as anonymous.
	Anonymous bool `json:"anonymous,omitempty"`
}

func EventToFieldMarshaling(evt *eth_abi.Event) (*FieldMarshaling, error) {
	var res = FieldMarshaling{
		Type:      "event",
		Name:      evt.RawName,
		Anonymous: evt.Anonymous,
	}

	for _, input := range evt.Inputs {
		arg, err := ArgumentToArgumentMarshaling(&input)

		if err != nil {
			return nil, err
		}

		res.Inputs = append(res.Inputs, arg)
	}

	return &res, nil
}

func MethodToFieldMarshaling(meth *eth_abi.Method) (*FieldMarshaling, error) {
	var res = FieldMarshaling{
		Type:            "function",
		Name:            meth.RawName,
		StateMutability: meth.StateMutability,
		Constant:        meth.Constant,
		Payable:         meth.Payable,
	}

	for _, input := range meth.Inputs {
		arg, err := ArgumentToArgumentMarshaling(&input)

		if err != nil {
			return nil, err
		}

		res.Inputs = append(res.Inputs, arg)
	}

	for _, output := range meth.Outputs {
		arg, err := ArgumentToArgumentMarshaling(&output)

		if err != nil {
			return nil, err
		}

		res.Outputs = append(res.Outputs, arg)
	}

	return &res, nil
}

func ArgumentToArgumentMarshaling(arg *eth_abi.Argument) (*ArgumentMarshaling, error) {
	m, err := TypeToArgumentMarshaling(arg.Type)

	if err != nil {
		return m, err
	}

	m.Name = arg.Name
	m.Indexed = arg.Indexed

	return m, nil
}

func TypeToArgumentMarshaling(t eth_abi.Type) (*ArgumentMarshaling, error) {
	switch t.T {
	case eth_abi.IntTy, eth_abi.UintTy, eth_abi.BoolTy, eth_abi.StringTy, eth_abi.AddressTy, eth_abi.FixedBytesTy, eth_abi.BytesTy, eth_abi.HashTy, eth_abi.FunctionTy:
		return &ArgumentMarshaling{
			Type:         t.String(),
			InternalType: t.String(),
		}, nil

	case eth_abi.ArrayTy, eth_abi.SliceTy:
		res, err := TypeToArgumentMarshaling(*t.Elem)

		if err != nil {
			return res, err
		}

		if t.Size == 0 {
			res.Type = res.Type + "[]"
			res.InternalType = res.InternalType + "[]"
		} else {
			res.Type = res.Type + "[" + strconv.FormatInt(int64(t.Size), 10) + "]"
			res.InternalType = res.InternalType + "[" + strconv.FormatInt(int64(t.Size), 10) + "]"
		}

		return res, nil

	case eth_abi.TupleTy:
		var res = ArgumentMarshaling{
			Type:         "tuple",
			InternalType: t.TupleRawName,
		}

		for i, comp := range t.TupleElems {
			m, err := TypeToArgumentMarshaling(*comp)

			if err != nil {
				return &res, err
			}

			m.Name = t.TupleRawNames[i]
			res.Components = append(res.Components, m)
		}

		return &res, nil

	default:
		return nil, fmt.Errorf("unhandled ABI type: %s", t.String())
	}
}
