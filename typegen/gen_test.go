package typegen

import (
	"fmt"
	"io/ioutil"
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
	fmt.Printf("%#v", tg.F)
	t.Log(g)
	t.Fail()
}

func TestGenBigMetadata(t *testing.T) {
	inp, err := ioutil.ReadFile("../json-gen/meta.json")
	require.NoError(t, err)
	mr, err := metadata.ParseMetadata(inp)

	tg := NewTypeGenerator(&mr, "typegen")
	res, err := tg.GenAll()

	ioutil.WriteFile("test_out.go", []byte(res), 0644)

	t.Fail()
}
