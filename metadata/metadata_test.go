package metadata

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicMetadata(t *testing.T) {
  raw, err := ioutil.ReadFile("../json-gen/meta.json")
  require.NoError(t, err)

  mr, err := ParseMetadata(raw)
  require.NoError(t, err)
  

  for _, tdef := range mr.Lookup.Types {
    _, err := tdef.Ty.GetTypeDef()
    assert.NoError(t, err)
  }
}
