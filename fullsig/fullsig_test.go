package fullsig

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

var (
	//go:embed fullsig_test_data.json
	fullsigTestDataJSON []byte
	fullsigTestData     []fullsigTestDataItem
)

type fullsigTestDataItem struct {
	Fullsig string
	Field   json.RawMessage
}

func init() {
	lo.Must0(json.Unmarshal(fullsigTestDataJSON, &fullsigTestData))
}
func TestFullsig(t *testing.T) {
	for _, item := range fullsigTestData {
		t.Run(item.Fullsig, func(t *testing.T) {
			var _abi, err = abi.JSON(bytes.NewReader(item.Field))
			assert.NoError(t, err)

			if len(_abi.Events) > 0 {
				var evt = lo.Values(_abi.Events)[0]
				res, err := ParseEvent(item.Fullsig)
				assert.NoError(t, err)
				assert.Equal(t, evt, res)
			} else {
				var meth = lo.Values(_abi.Methods)[0]
				res, err := ParseMethod(item.Fullsig)
				assert.NoError(t, err)
				assert.Equal(t, meth, res)
			}
		})
	}
}
