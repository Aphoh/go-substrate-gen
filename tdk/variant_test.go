package tdk

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicVariant(t *testing.T) {
	vari := TDVariant{}
	err := json.Unmarshal([]byte(testVariantJson), &vari)
	assert.NoError(t, err)
}

const testVariantJson = `
{
  "variants": [
    {
      "name": "ChangesTrieRoot",
      "fields": [
        {
          "name": null,
          "type": "9",
          "typeName": "Hash",
          "docs": []
        }
      ],
      "index": "2",
      "docs": []
    },
    {
      "name": "PreRuntime",
      "fields": [
        {
          "name": null,
          "type": "14",
          "typeName": "ConsensusEngineId",
          "docs": []
        },
        {
          "name": null,
          "type": "10",
          "typeName": "Vec<u8>",
          "docs": []
        }
      ],
      "index": "6",
      "docs": []
    },
    {
      "name": "Consensus",
      "fields": [
        {
          "name": null,
          "type": "14",
          "typeName": "ConsensusEngineId",
          "docs": []
        },
        {
          "name": null,
          "type": "10",
          "typeName": "Vec<u8>",
          "docs": []
        }
      ],
      "index": "4",
      "docs": []
    },
    {
      "name": "Seal",
      "fields": [
        {
          "name": null,
          "type": "14",
          "typeName": "ConsensusEngineId",
          "docs": []
        },
        {
          "name": null,
          "type": "10",
          "typeName": "Vec<u8>",
          "docs": []
        }
      ],
      "index": "5",
      "docs": []
    },
    {
      "name": "ChangesTrieSignal",
      "fields": [
        {
          "name": null,
          "type": "15",
          "typeName": "ChangesTrieSignal",
          "docs": []
        }
      ],
      "index": "7",
      "docs": []
    },
    {
      "name": "Other",
      "fields": [
        {
          "name": null,
          "type": "10",
          "typeName": "Vec<u8>",
          "docs": []
        }
      ],
      "index": "0",
      "docs": []
    },
    {
      "name": "RuntimeEnvironmentUpdated",
      "fields": [],
      "index": "8",
      "docs": []
    }
  ]
}
`
