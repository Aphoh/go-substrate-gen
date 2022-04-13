package palletgen

import (
	"io/ioutil"
	"testing"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/stretchr/testify/require"
)

// TODO: better method for testing
func TestGenBigMetadata(t *testing.T) {
	inp, err := ioutil.ReadFile("../json-gen/meta.json")
	require.NoError(t, err)
	mr, err := metadata.ParseMetadata(inp)

	for p := range mr.Pallets {
		tg := typegen.NewTypeGenerator(&mr, "github.com/aphoh/go-substrate-gen/palletgen")
		palletGen := NewPalletGenerator(&mr.Pallets[p], &tg)

		_, _, err := palletGen.GenerateStorage("github.com/aphoh/go-substrate-gen/palletgen")
		require.NoError(t, err)

		_, _, err = palletGen.GenerateCalls("github.com/aphoh/go-substrate-gen/palletgen")
		require.NoError(t, err)
	}
}
