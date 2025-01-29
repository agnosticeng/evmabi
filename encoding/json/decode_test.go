package json

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"testing"

	eth_abi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/swaggest/assertjson"

	"github.com/samber/lo"
)

var (
	//go:embed abi.json
	abiStr []byte
	_abi   = lo.Must(eth_abi.JSON(bytes.NewReader(abiStr)))

	//go:embed trace_test_data.json
	traceTestDataJSON []byte
	traceTestData     []traceTestDataItem

	//go:embed log_test_data.json
	logTestDataJSON []byte
	logTestData     []logTestDataItem
)

type traceTestDataItem struct {
	MethodName string
	Input      string
	Output     string
	Result     json.RawMessage
}

type logTestDataItem struct {
	EventName string
	Input     string
	Topics    []string
	Result    json.RawMessage
}

func init() {
	lo.Must0(json.Unmarshal(traceTestDataJSON, &traceTestData))
	lo.Must0(json.Unmarshal(logTestDataJSON, &logTestData))
}

func TestDecodeTrace(t *testing.T) {
	for _, trace := range traceTestData {
		t.Run(trace.MethodName, func(t *testing.T) {
			var (
				method = _abi.Methods[trace.MethodName]
				input  = hexutil.MustDecode(trace.Input)
				output []byte
			)

			if len(trace.Output) > 0 {
				output = hexutil.MustDecode(trace.Output)
			}

			node, err := DecodeTrace(input, output, method)
			assert.NoError(t, err)
			js, err := node.MarshalJSON()
			assert.NoError(t, err)
			assertjson.Equal(t, []byte(trace.Result), js)
		})
	}
}

func TestDecodeLog(t *testing.T) {
	for _, log := range logTestData {
		t.Run(log.EventName, func(t *testing.T) {
			var (
				event  = _abi.Events[log.EventName]
				input  = hexutil.MustDecode(log.Input)
				topics = lo.Map(log.Topics, func(topic string, _ int) [32]byte { return [32]byte(hexutil.MustDecode(topic)) })
			)

			node, err := DecodeLog(topics, input, event)
			assert.NoError(t, err)
			js, err := node.MarshalJSON()
			assert.NoError(t, err)
			assertjson.Equal(t, []byte(log.Result), js)
		})
	}
}
