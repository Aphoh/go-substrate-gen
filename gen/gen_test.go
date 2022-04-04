package gen

import (
	"fmt"
	"testing"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJson = `
{
  "lookup": {
    "types": [
      {
        "id": "0",
        "type": {
          "path": [
            "sp_core",
            "crypto",
            "AccountId32"
          ],
          "params": [],
          "def": {
            "Composite": {
              "fields": [
                {
                  "name": null,
                  "type": "1",
                  "typeName": "[u8; 32]",
                  "docs": []
                }
              ]
            }
          },
          "docs": []
        }
      },
      {
        "id": "1",
        "type": {
          "path": [],
          "params": [],
          "def": {
            "Array": {
              "len": "32",
              "type": "2"
            }
          },
          "docs": []
        }
      },
      {
        "id": "2",
        "type": {
          "path": [],
          "params": [],
          "def": {
            "Primitive": "U8"
          },
          "docs": []
        }
      }
    ]
  }
}
`

func TestGenSmallMetadata(t *testing.T) {
	mr, err := metadata.ParseMetadata([]byte(testJson))
	require.NoError(t, err)
	tg := NewTypeGenerator(&mr, "example")

	g, err := tg.GetType("0")
	assert.NoError(t, err)
	fmt.Printf("%#v", tg.f)
	t.Log(g)
	t.Fail()
}
