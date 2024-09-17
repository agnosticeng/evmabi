package fullsig

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/wolfeidau/stringtokenizer"
)

var zero eth_abi.ArgumentMarshaling

func ParseEvent(s string) (eth_abi.Event, error) {
	s = strings.ReplaceAll(s, " ", "")

	var (
		tokenizer = stringtokenizer.NewStringTokenizer(
			strings.NewReader(s),
			"(),[]",
			true,
		)
	)

	name, err := readIdentifier(tokenizer)

	if err != nil {
		return eth_abi.Event{}, err
	}

	if err := readLeftParens(tokenizer); err != nil {
		return eth_abi.Event{}, err
	}

	inputs, err := parseArguments(tokenizer)

	if err != nil {
		return eth_abi.Event{}, err
	}

	if err := readEOF(tokenizer); err != nil {
		return eth_abi.Event{}, err
	}

	return eth_abi.NewEvent(name, name, false, inputs), nil
}

func ParseMethod(s string) (eth_abi.Method, error) {
	s = strings.ReplaceAll(s, " ", "")

	var (
		tokenizer = stringtokenizer.NewStringTokenizer(
			strings.NewReader(s),
			"(),[]",
			true,
		)
	)

	name, err := readIdentifier(tokenizer)

	if err != nil {
		return eth_abi.Method{}, err
	}

	if err := readLeftParens(tokenizer); err != nil {
		return eth_abi.Method{}, err
	}

	inputs, err := parseArguments(tokenizer)

	if err != nil {
		return eth_abi.Method{}, err
	}

	t, err := readToken(tokenizer)

	if errors.Is(err, io.EOF) {
		return eth_abi.NewMethod(
			t,
			t,
			eth_abi.Function,
			"",
			false,
			false,
			inputs,
			nil,
		), nil
	}

	if err != nil {
		return eth_abi.Method{}, err
	}

	if t != "(" {
		return eth_abi.Method{}, fmt.Errorf("unattended token: %s (wanted '(')", t)
	}

	outputs, err := parseArguments(tokenizer)

	if err != nil {
		return eth_abi.Method{}, err
	}

	if err := readEOF(tokenizer); err != nil {
		return eth_abi.Method{}, err
	}

	return eth_abi.NewMethod(
		name,
		name,
		eth_abi.Function,
		"",
		false,
		false,
		inputs,
		outputs,
	), nil
}

func parseArguments(tokenizer *stringtokenizer.StringTokenizer) (eth_abi.Arguments, error) {
	tuple, err := parseTuple(tokenizer)

	if err != nil {
		return nil, err
	}

	var args eth_abi.Arguments

	for _, arg := range tuple.Components {
		argType, err := eth_abi.NewType(arg.Type, arg.InternalType, arg.Components)

		if err != nil {
			return nil, err
		}

		args = append(args, eth_abi.Argument{
			Type:    argType,
			Indexed: arg.Indexed,
			Name:    arg.Name,
		})
	}

	return args, nil
}

func parseTuple(tokenizer *stringtokenizer.StringTokenizer) (eth_abi.ArgumentMarshaling, error) {
	var (
		res eth_abi.ArgumentMarshaling
		i   int
	)

	for {
		arg, err := parseItem(tokenizer)

		if err != nil {
			return zero, err
		}

		arg.Name = fmt.Sprintf("arg%d", i)

		t, err := readToken(tokenizer)

		if err != nil {
			return zero, err
		}

		if t == "[" {
			length, err := readArrayLength(tokenizer)

			if err != nil {
				return zero, err
			}

			if length == 0 {
				arg.Type = arg.Type + "[]"
			} else {
				arg.Type = arg.Type + fmt.Sprintf("[%d]", length)
			}

			t, err = readToken(tokenizer)

			if err != nil {
				return zero, err
			}
		}

		res.Components = append(res.Components, arg)

		if t == ")" {
			break
		}

		if t != "," {
			return zero, fmt.Errorf("unattended token: %s (wanted ',')", t)
		}

		i++
	}

	res.Type = "tuple"
	return res, nil
}

func parseItem(tokenizer *stringtokenizer.StringTokenizer) (eth_abi.ArgumentMarshaling, error) {
	var (
		res eth_abi.ArgumentMarshaling
		err error
	)

	t, err := readToken(tokenizer)

	if err != nil {
		return zero, err
	}

	t, found := strings.CutPrefix(t, "indexed")

	switch {
	case t == "address":
		res = eth_abi.ArgumentMarshaling{Type: t}

	case t == "bool":
		res = eth_abi.ArgumentMarshaling{Type: t}

	case t == "string":
		res = eth_abi.ArgumentMarshaling{Type: t}

	case strings.HasPrefix(t, "uint"):
		res = eth_abi.ArgumentMarshaling{Type: t}

	case strings.HasPrefix(t, "int"):
		res = eth_abi.ArgumentMarshaling{Type: t}

	case strings.HasPrefix(t, "bytes"):
		res = eth_abi.ArgumentMarshaling{Type: t}

	case t == "(":
		res, err = parseTuple(tokenizer)

	default:
		return zero, fmt.Errorf("unattended token: %s", t)
	}

	if err != nil {
		return zero, err
	}

	if found {
		res.Indexed = true
	}

	return res, nil
}

func readArrayLength(tokenizer *stringtokenizer.StringTokenizer) (int, error) {
	t, err := readToken(tokenizer)

	if err != nil {
		return 0, err
	}

	if t == "]" {
		return 0, nil
	}

	i, err := strconv.ParseInt(t, 10, 64)

	if err != nil {
		return 0, err
	}

	t, err = readToken(tokenizer)

	if err != nil {
		return 0, err
	}

	if t != "]" {
		return 0, fmt.Errorf("unattended token: %s (wanted ']')", t)
	}

	return int(i), nil
}

func readIdentifier(tokenizer *stringtokenizer.StringTokenizer) (string, error) {
	identifier, err := readToken(tokenizer)

	if err != nil {
		return "", nil
	}

	if !isIdentifier(identifier) {
		return "", fmt.Errorf("unattended token: %s", identifier)
	}

	return identifier, nil
}

func readLeftParens(tokenizer *stringtokenizer.StringTokenizer) error {
	parens, err := readToken(tokenizer)

	if err != nil {
		return err
	}

	if parens != "(" {
		return fmt.Errorf("unattended token: %s (wanted '(')", parens)
	}

	return nil
}

func readRightParens(tokenizer *stringtokenizer.StringTokenizer) error {
	parens, err := readToken(tokenizer)

	if err != nil {
		return err
	}

	if parens != ")" {
		return fmt.Errorf("unattended token: %s (wanted ')')", parens)
	}

	return nil
}

func readEOF(tokenizer *stringtokenizer.StringTokenizer) error {
	t, err := readToken(tokenizer)

	if errors.Is(err, io.EOF) {
		return nil
	}

	return fmt.Errorf("unattended token: %s (wanted EOF)')", t)
}

func readToken(tokenizer *stringtokenizer.StringTokenizer) (string, error) {
	if tokenizer.HasMoreTokens() {
		return strings.ReplaceAll(tokenizer.NextToken(), " ", ""), nil
	}

	return "", io.EOF
}

func isIdentifier(value string) bool {
	for _, c := range value {
		if !unicode.IsDigit(c) && !unicode.IsLetter(c) {
			return false
		}
	}

	return true
}
