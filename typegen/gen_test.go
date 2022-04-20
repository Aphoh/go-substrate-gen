package typegen

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/stretchr/testify/require"
)

// TODO: better method for testing
func TestGenBigMetadata(t *testing.T) {
	inp, err := ioutil.ReadFile("../polkadot-meta.json")
	require.NoError(t, err)
	mr, encMeta, err := metadata.ParseMetadata(inp)
	require.NoError(t, err)

	tg := NewTypeGenerator(mr, encMeta, "github.com/aphoh/go-substrate-gen/typegen")
	res, err := tg.GenAll()

	require.False(t, strings.Contains(res, "%!v(PANIC="), "Generated code contains errors")
	//ioutil.WriteFile("test_out.go", []byte(res), 0644)
}
