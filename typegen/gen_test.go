package typegen

import (
	"io/ioutil"
	"strings"
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
                  "name": "a",
                  "type": "1",
                  "typeName": "[u8; 32]",
                  "docs": []
                },
                {
                  "name": "b",
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
	t.Log(mr)
	require.NoError(t, err)
	tg := NewTypeGenerator(&mr, "github.com/aphoh/go-substrate-gen/typegen")

	res, err := tg.GenAll()
	assert.NoError(t, err)

	//ioutil.WriteFile("test_out.go", []byte(res), 0644)
	_ = res

	t.Fail()
}

// TODO: better method for testing
func TestGenBigMetadata(t *testing.T) {
	inp, err := ioutil.ReadFile("../json-gen/meta.json")
	require.NoError(t, err)
	mr, err := metadata.ParseMetadata(inp)

	tg := NewTypeGenerator(&mr, "github.com/aphoh/go-substrate-gen/typegen")
	res, err := tg.GenAll()

	require.False(t, strings.Contains(res, "%!v(PANIC="), "Generated code contains errors")

	ioutil.WriteFile("test_out.go", []byte(res), 0644)
}
