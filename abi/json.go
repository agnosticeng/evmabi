package abi

import (
	"encoding/json"
	"fmt"

	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
)

type field struct {
	Type    string
	Name    string
	Inputs  []eth_abi.Argument
	Outputs []eth_abi.Argument

	// Status indicator which can be: "pure", "view",
	// "nonpayable" or "payable".
	StateMutability string

	// Deprecated Status indicators, but removed in v0.6.0.
	Constant bool // True if function is either pure or view
	Payable  bool // True if function is payable

	// Event relevant indicator represents the event is
	// declared as anonymous.
	Anonymous bool
}

func JSONEvent(data []byte) (*eth_abi.Event, error) {
	var field field

	if err := json.Unmarshal(data, &field); err != nil {
		return nil, err
	}

	if len(field.Name) == 0 {
		return nil, fmt.Errorf("field descriptor name must not be empty")
	}

	if field.Type != "event" {
		return nil, fmt.Errorf("wrong field type: %s", field.Type)
	}

	var evt = eth_abi.NewEvent(field.Name, field.Name, field.Anonymous, field.Inputs)
	return &evt, nil
}

func JSONMethod(data []byte) (*eth_abi.Method, error) {
	var field field

	if err := json.Unmarshal(data, &field); err != nil {
		return nil, err
	}

	if len(field.Name) == 0 {
		return nil, fmt.Errorf("field descriptor name must not be empty")
	}

	if field.Type != "function" {
		return nil, fmt.Errorf("wrong field type: %s", field.Type)
	}

	var meth = eth_abi.NewMethod(field.Name, field.Name, eth_abi.Function, field.StateMutability, field.Constant, field.Payable, field.Inputs, field.Outputs)
	return &meth, nil
}
