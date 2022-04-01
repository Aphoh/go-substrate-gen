package tdk

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testArrayJson = `
{
  "Array": {
    "len": "32",
    "type": "2"
  }
}
`

func TestArrayBasic(t *testing.T) {
	arr := TDArray{}
	err := json.Unmarshal([]byte(testArrayJson), &arr)
	assert.NoError(t, err)
}
