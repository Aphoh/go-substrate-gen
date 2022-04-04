package tdk

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const compositeJson = `
{
  "fields": [
    {
      "name": null,
      "type": "1",
      "typeName": "[u8; 32]",
      "docs": []
    }
  ]
}`

func TestCompositeParsesBasic(t *testing.T) {
	comp := TDComposite{}
	err := json.Unmarshal([]byte(compositeJson), &comp)
	assert.NoError(t, err)

	assert.Equal(t, comp, TDComposite{
		Fields: []TDField{
			{
				Name:     "",
				TypeId:   "1",
				TypeName: "[u8; 32]",
				Docs:     []string{},
			},
		},
	})
}
