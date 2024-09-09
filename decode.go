package evmabi

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"unicode/utf8"

	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

func decodeArguments(
	data []byte,
	args eth_abi.Arguments,
	fn YieldFunc,
) error {
	var virtualArgs = 0

	if err := Yield(fn, &Event{
		Type: TupleStart,
		Len:  len(args),
	}); err != nil {
		return err
	}

	for i, arg := range args {
		if err := Yield(fn, &Event{
			Type:  Key,
			Key:   arg.Name,
			Index: i,
		}); err != nil {
			return err
		}

		if err := decodeArgument(
			data,
			arg,
			(i+virtualArgs)*32,
			fn,
		); err != nil {
			return err
		}

		if arg.Type.T == eth_abi.ArrayTy && !isDynamic(arg.Type) {
			virtualArgs += typeSize(arg.Type)/32 - 1
		} else if arg.Type.T == eth_abi.TupleTy && !isDynamic(arg.Type) {
			virtualArgs += typeSize(arg.Type)/32 - 1
		}
	}

	if err := Yield(fn, &Event{
		Type: TupleEnd,
	}); err != nil {
		return err
	}

	return nil
}

func decodeArgument(
	data []byte,
	arg eth_abi.Argument,
	idx int,
	fn YieldFunc,
) error {
	if isDynamic(arg.Type) && arg.Indexed {
		return ErrDynamicIndexedArgument
	}

	if err := decodeValue(
		data,
		arg.Type,
		idx,
		fn,
	); err != nil {
		return err
	}

	return nil
}

func decodeValue(
	data []byte,
	t eth_abi.Type,
	idx int,
	fn YieldFunc,
) error {
	if idx+32 > len(data) {
		return fmt.Errorf("idx points over data slice boundary")
	}

	var (
		returnOutput  []byte
		begin, length int
		err           error
	)

	if isLengthPrefixed(t) {
		begin, length, err = decodeLengthPrefix(data, idx)

		if err != nil {
			return err
		}
	} else {
		returnOutput = data[idx : idx+32]
	}

	switch t.T {
	case eth_abi.UintTy:
		var i = uint256.NewInt(0).SetBytes(returnOutput)

		if i.BitLen() > t.Size {
			return fmt.Errorf("uint needs too many bits (%d/%d)", i.BitLen(), t.Size)
		}

		return Yield(fn, &Event{
			Type:    Value,
			ABIType: t,
			Value:   i,
		})

	case eth_abi.IntTy:
		var i = uint256.NewInt(0).SetBytes(returnOutput)

		if i.BitLen() > t.Size {
			return fmt.Errorf("int needs too many bits (%d/%d)", i.BitLen(), t.Size)
		}

		return Yield(fn, &Event{
			Type:    Value,
			ABIType: t,
			Value:   i,
		})

	case eth_abi.BoolTy:
		var b, err = readBool(returnOutput)

		if err != nil {
			return err
		}

		return Yield(fn, &Event{
			Type:    Value,
			ABIType: t,
			Value:   b,
		})

	case eth_abi.AddressTy:
		return Yield(fn, &Event{
			Type:    Value,
			ABIType: t,
			Value:   common.BytesToAddress(returnOutput[12:]),
		})

	case eth_abi.HashTy:
		return Yield(fn, &Event{
			Type:    Value,
			ABIType: t,
			Value:   common.BytesToHash(returnOutput),
		})

	case eth_abi.StringTy:
		var v = string(data[begin : begin+length])

		if !utf8.ValidString(v) {
			v = strconv.Quote(v)
		}

		return Yield(fn, &Event{
			Type:    Value,
			ABIType: t,
			Value:   v,
		})

	case eth_abi.BytesTy:
		return Yield(fn, &Event{
			Type:    Value,
			ABIType: t,
			Value:   common.CopyBytes(data[begin : begin+length]),
		})

	case eth_abi.FixedBytesTy:
		return Yield(fn, &Event{
			Type:    Value,
			ABIType: t,
			Value:   common.CopyBytes(returnOutput[0:t.Size]),
		})

	case eth_abi.FunctionTy:
		return Yield(fn, &Event{
			Type:    Value,
			ABIType: t,
			Value:   common.CopyBytes(returnOutput[0:t.Size]),
		})

	case eth_abi.TupleTy:
		if isDynamic(t) {
			offset, overflow := uint256.NewInt(0).SetBytes(data[idx : idx+32]).Uint64WithOverflow()

			if overflow {
				return fmt.Errorf("offset larger than uint64")
			}

			return decodeTuple(data[offset:], t, fn)
		}

		return decodeTuple(data[idx:], t, fn)

	case eth_abi.ArrayTy:
		if isDynamic(*t.Elem) {
			var offset = binary.BigEndian.Uint64(returnOutput[len(returnOutput)-8:])

			if offset > uint64(len(data)) {
				return fmt.Errorf("offset greater than data length")
			}

			return decodeArray(data[offset:], t, fn, 0, t.Size)
		}

		return decodeArray(data[idx:], t, fn, 0, t.Size)

	case eth_abi.SliceTy:
		return decodeArray(data[begin:], t, fn, 0, length)

	default:
		return fmt.Errorf("abi: unknown type %v", t.T)
	}
}

func decodeTuple(
	data []byte,
	t eth_abi.Type,
	fn YieldFunc,
) error {
	var virtualArgs = 0

	if err := Yield(fn, &Event{
		Type:    TupleStart,
		ABIType: t,
		Len:     len(t.TupleElems),
	}); err != nil {
		return err
	}

	for i, elem := range t.TupleElems {
		if err := Yield(fn, &Event{
			Type:  Key,
			Key:   t.TupleRawNames[i],
			Index: i,
		}); err != nil {
			return err
		}

		if err := decodeValue(
			data,
			*elem,
			(i+virtualArgs)*32,
			fn,
		); err != nil {
			return err
		}

		if elem.T == eth_abi.ArrayTy && !isDynamic(*elem) {
			virtualArgs += typeSize(*elem)/32 - 1
		} else if elem.T == eth_abi.TupleTy && !isDynamic(*elem) {
			virtualArgs += typeSize(*elem)/32 - 1
		}
	}

	if err := Yield(fn, &Event{
		Type: TupleEnd,
	}); err != nil {
		return err
	}

	return nil
}

func decodeArray(
	data []byte,
	t eth_abi.Type,
	fn YieldFunc,
	start int,
	size int,
) error {
	if size < 0 {
		return fmt.Errorf("cannot marshal input to array, size is negative (%d)", size)
	}

	if start+32*size > len(data) {
		return fmt.Errorf("abi: cannot marshal into go array: offset %d would go over slice boundary (len=%d)", len(data), start+32*size)
	}

	var elemSize = typeSize(*t.Elem)

	if err := Yield(fn, &Event{
		Type:    ArrayStart,
		ABIType: t,
		Len:     size,
	}); err != nil {
		return err
	}

	for i := 0; i < size; i++ {
		var index = start + (i * elemSize)

		if err := decodeValue(
			data,
			*t.Elem,
			index,
			fn,
		); err != nil {
			return err
		}
	}

	if err := Yield(fn, &Event{
		Type: ArrayEnd,
	}); err != nil {
		return err
	}

	return nil
}

func readBool(data []byte) (bool, error) {
	for _, b := range data[:31] {
		if b != 0 {
			return false, fmt.Errorf("invalid bool")
		}
	}

	switch data[31] {
	case 0:
		return false, nil
	case 1:
		return true, nil

	default:
		return false, fmt.Errorf("invalid bool")
	}
}
