package encoding

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/holiman/uint256"
	"github.com/samber/lo"
)

func geteDataSlice(data []byte, idx int, t eth_abi.Type) ([]byte, error) {
	if isLengthPrefixed(t) {
		offset, length, err := decodeLengthPrefix(data, idx)

		if err != nil {
			return nil, err
		}

		return data[offset : offset+length], nil
	} else {
		return data[idx : idx+32], nil
	}
}

func decodeLengthPrefix(data []byte, idx int) (int, int, error) {
	offset, overflow := uint256.NewInt(0).SetBytes(data[idx : idx+32]).Uint64WithOverflow()

	if overflow {
		return 0, 0, fmt.Errorf("offset larger than uint64")
	}

	if (offset + 32) > uint64(len(data)) {
		return 0, 0, fmt.Errorf("offset points over data slice boundary")
	}

	length, overflow := uint256.NewInt(0).SetBytes(data[offset : offset+32]).Uint64WithOverflow()

	if overflow {
		return 0, 0, fmt.Errorf("length larger than uint64")
	}

	if (offset + 32 + length) > uint64(len(data)) {
		return 0, 0, fmt.Errorf("offset+length points over data slice boundary")
	}

	return int(offset) + 32, int(length), nil
}

func isLengthPrefixed(t eth_abi.Type) bool {
	return t.T == eth_abi.StringTy || t.T == eth_abi.BytesTy || t.T == eth_abi.SliceTy
}

func isDynamic(t eth_abi.Type) bool {
	switch t.T {
	case eth_abi.StringTy, eth_abi.BytesTy, eth_abi.SliceTy:
		return true

	case eth_abi.TupleTy:
		for _, elem := range t.TupleElems {
			if isDynamic(*elem) {
				return true
			}
		}

		return false

	case eth_abi.ArrayTy:
		return isDynamic(*t.Elem)

	case eth_abi.HashTy, eth_abi.AddressTy, eth_abi.BoolTy, eth_abi.IntTy, eth_abi.UintTy, eth_abi.FixedBytesTy:
		return false

	default:
		panic(fmt.Sprintf("cannot determine if this type is dynamic: %v", t.String()))
	}
}

func typeSize(t eth_abi.Type) int {
	switch {
	case t.T == eth_abi.ArrayTy && !isDynamic(t):
		return t.Size * typeSize(*t.Elem)

	case t.T == eth_abi.TupleTy && isDynamic(t):
		return lo.SumBy(t.TupleElems, func(t *eth_abi.Type) int { return typeSize(*t) })

	default:
		return 32
	}
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

func isValidFieldName(fieldName string) bool {
	for i, c := range fieldName {
		if i == 0 && !isLetter(c) {
			return false
		}

		if !(isLetter(c) || unicode.IsDigit(c)) {
			return false
		}
	}

	return len(fieldName) > 0
}
