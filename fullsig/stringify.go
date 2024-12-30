package fullsig

import (
	"strconv"
	"strings"

	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
)

func StringifyEvent(evt *eth_abi.Event) string {
	var sb strings.Builder

	sb.WriteString("event ")
	sb.WriteString(evt.RawName)
	sb.WriteString("(")

	for i, input := range evt.Inputs {
		sb.WriteString(StringifyArgument(&input))

		if i != (len(evt.Inputs) - 1) {
			sb.WriteString(",")
		}
	}

	sb.WriteString(")")
	return sb.String()
}

func StringifyMethod(meth *eth_abi.Method) string {
	var sb strings.Builder

	sb.WriteString("function ")
	sb.WriteString(meth.RawName)
	sb.WriteString("(")

	for i, input := range meth.Inputs {
		sb.WriteString(StringifyArgument(&input))

		if i != (len(meth.Inputs) - 1) {
			sb.WriteString(",")
		}
	}

	sb.WriteString(")")

	if len(meth.Outputs) > 0 {
		sb.WriteString("(")

		for i, output := range meth.Outputs {
			sb.WriteString(StringifyArgument(&output))

			if i != (len(meth.Outputs) - 1) {
				sb.WriteString(",")
			}
		}

		sb.WriteString(")")
	}

	return sb.String()
}

func StringifyArgument(arg *eth_abi.Argument) string {
	var sb strings.Builder

	sb.WriteString(StringifyType(&arg.Type))

	if arg.Indexed {
		sb.WriteString(" indexed")
	}

	return sb.String()
}

func StringifyType(t *eth_abi.Type) string {
	switch t.T {
	case eth_abi.SliceTy, eth_abi.ArrayTy:
		var sb strings.Builder
		sb.WriteString(StringifyType(t.Elem))
		sb.WriteString("[")

		if t.Size > 0 {

			sb.WriteString(strconv.FormatInt(int64(t.Size), 10))
		}

		sb.WriteString("]")
		return sb.String()

	case eth_abi.TupleTy:
		var sb strings.Builder
		sb.WriteString("(")

		for i := 0; i < len(t.TupleRawNames); i++ {
			sb.WriteString(StringifyType(t.TupleElems[i]))

			if i < len(t.TupleRawNames)-1 {
				sb.WriteString(",")
			}
		}

		sb.WriteString(")")
		return sb.String()

	default:
		return t.String()
	}
}
