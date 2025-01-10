package fullsig

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/bzick/tokenizer"
	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/samber/lo"
)

const (
	TokenOpenParens   = 1
	TokenCloseParens  = 2
	TokenOpenBracket  = 3
	TokenCloseBracket = 4
	TokenComma        = 5
)

var (
	SCALAR_TYPENAMES = lo.Flatten([][]string{
		{
			"address",
			"bool",
			"string",
			"bytes",
			"uint",
			"function",
		},
		lo.RepeatBy(32, func(i int) string {
			return "uint" + strconv.FormatUint(uint64((i+1)*8), 10)
		}),
		lo.RepeatBy(32, func(i int) string {
			return "int" + strconv.FormatUint(uint64((i+1)*8), 10)
		}),
		lo.RepeatBy(32, func(i int) string {
			return "bytes" + strconv.FormatUint(uint64((i+1)), 10)
		}),
	})

	tknz = tokenizer.New().
		DefineTokens(TokenOpenParens, []string{"("}).
		DefineTokens(TokenCloseParens, []string{")"}).
		DefineTokens(TokenOpenBracket, []string{"["}).
		DefineTokens(TokenCloseBracket, []string{"]"}).
		DefineTokens(TokenComma, []string{","}).
		AllowKeywordSymbols([]rune{'_'}, []rune{'$', '_', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'})
)

func isEventToken(t *tokenizer.Token) bool    { return t.IsKeyword() && t.ValueString() == "event" }
func isFunctionToken(t *tokenizer.Token) bool { return t.IsKeyword() && t.ValueString() == "function" }
func isIndexedToken(t *tokenizer.Token) bool  { return t.IsKeyword() && t.ValueString() == "indexed" }
func isScalarTypeName(t *tokenizer.Token) bool {
	return lo.Contains(SCALAR_TYPENAMES, t.ValueString())
}

func ParseEvent(s string) (eth_abi.Event, error) {
	var stream = tknz.ParseString(s)
	defer stream.Close()

	if !isEventToken(stream.CurrentToken()) {
		return eth_abi.Event{}, fmt.Errorf("event fullsig must start with keyword 'event': %s", s)
	}

	if !stream.GoNext().CurrentToken().IsKeyword() {
		return eth_abi.Event{}, fmt.Errorf("wanted event name but got: %s", string(stream.CurrentToken().Value()))
	}

	var eventName = stream.CurrentToken().ValueString()
	stream.GoNext()
	inputs, err := newArguments(stream)

	if err != nil {
		return eth_abi.Event{}, err
	}

	if !stream.GoNext().IsValid() {
		return eth_abi.Event{}, fmt.Errorf("wanted EOF but got %s", string(stream.CurrentToken().Value()))
	}

	return eth_abi.NewEvent(eventName, eventName, false, inputs), nil
}

func ParseMethod(s string) (eth_abi.Method, error) {
	var stream = tknz.ParseString(s)
	defer stream.Close()

	if !isFunctionToken(stream.CurrentToken()) {
		return eth_abi.Method{}, fmt.Errorf("function fullsig must start with keyword 'function': %s", s)
	}

	if !stream.GoNext().CurrentToken().IsKeyword() {
		return eth_abi.Method{}, fmt.Errorf("wanted function name but got: %s", string(stream.CurrentToken().Value()))
	}

	var functionName = stream.CurrentToken().ValueString()
	stream.GoNext()
	inputs, err := newArguments(stream)

	if err != nil {
		return eth_abi.Method{}, err
	}

	var outputs eth_abi.Arguments

	if stream.CurrentToken().Is(TokenOpenParens) {
		outputs, err = newArguments(stream)

		if err != nil {
			return eth_abi.Method{}, err
		}
	}

	if !stream.GoNext().IsValid() {
		return eth_abi.Method{}, fmt.Errorf("wanted EOF but got %s", string(stream.CurrentToken().Value()))
	}

	return eth_abi.NewMethod(functionName, functionName, eth_abi.Function, "", false, false, inputs, outputs), nil
}

func newArguments(stream *tokenizer.Stream) (eth_abi.Arguments, error) {
	args, err := parseArguments(stream)

	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		return nil, nil
	}

	var res = make([]eth_abi.Argument, len(args))

	for i, arg := range args {
		t, err := eth_abi.NewType(arg.Type, arg.InternalType, arg.Components)

		if err != nil {
			js, _ := json.MarshalIndent(arg.Components, "", "    ")
			fmt.Println(string(js))
			return nil, err
		}

		var argRes eth_abi.Argument
		argRes.Type = t
		argRes.Indexed = arg.Indexed
		argRes.Name = arg.Name

		res[i] = argRes
	}

	return res, nil
}

func parseArgument(stream *tokenizer.Stream) (eth_abi.ArgumentMarshaling, error) {
	var (
		res eth_abi.ArgumentMarshaling
		err error
	)

	switch {
	case isScalarTypeName(stream.CurrentToken()):
		res.Type = stream.CurrentToken().ValueString()
		stream.GoNext()
	case stream.CurrentToken().Is(TokenOpenParens):
		res.Type = "tuple"
		res.Components, err = parseArguments(stream)
	default:
		return res, fmt.Errorf("invalid token: %s", string(stream.CurrentToken().Value()))
	}

	if err != nil {
		return res, err
	}

	for stream.CurrentToken().Is(TokenOpenBracket) {
		s, err := parseArraySuffix(stream)

		if err != nil {
			return res, err
		}

		res.Type = res.Type + s
	}

	if isIndexedToken(stream.CurrentToken()) {
		res.Indexed = true
		stream.GoNext()
	}

	return res, nil
}

func parseArraySuffix(stream *tokenizer.Stream) (string, error) {
	if !stream.CurrentToken().Is(TokenOpenBracket) {
		return "", fmt.Errorf("wanted token '[' but got %s", string(stream.CurrentToken().Value()))
	}

	stream.GoNext()

	switch {
	case stream.CurrentToken().Is(TokenCloseBracket):
		stream.GoNext()
		return "[]", nil
	case stream.CurrentToken().Is(tokenizer.TokenInteger) && stream.NextToken().Is(TokenCloseBracket):
		var i = stream.CurrentToken().ValueInt64()
		stream.GoNext().GoNext()
		return "[" + strconv.FormatInt(i, 10) + "]", nil

	default:
		return "", fmt.Errorf("wanted token ']' or integer then ']' but got %s", string(stream.CurrentToken().Value()))
	}
}

func parseArguments(stream *tokenizer.Stream) ([]eth_abi.ArgumentMarshaling, error) {
	var res []eth_abi.ArgumentMarshaling

	if !stream.CurrentToken().Is(TokenOpenParens) {
		return nil, fmt.Errorf("wanted token '(' but got %s", string(stream.CurrentToken().Value()))
	}

	stream.GoNext()

	if stream.CurrentToken().Is(TokenCloseParens) {
		stream.GoNext()
		return res, nil
	}

	for {
		arg, err := parseArgument(stream)

		if err != nil {
			return nil, err
		}

		if len(arg.Name) == 0 {
			arg.Name = "arg" + strconv.FormatInt(int64(len(res)), 10)
		}

		res = append(res, arg)

		switch {
		case stream.CurrentToken().Is(TokenComma):
			stream.GoNext()
		case stream.CurrentToken().Is(TokenCloseParens):
			stream.GoNext()
			return res, nil
		default:
			return nil, fmt.Errorf("wanted token ',' or ')' but got %s", string(stream.CurrentToken().Value()))
		}
	}
}
