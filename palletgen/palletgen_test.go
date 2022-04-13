package palletgen

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/stretchr/testify/require"
)

// TODO: better method for testing
func TestGenBigMetadata(t *testing.T) {
	inp, err := ioutil.ReadFile("../polka-meta.json")
	require.NoError(t, err)
	mr, encMeta, err := metadata.ParseMetadata(inp)

	for _, pal := range mr.Pallets {
		tg := typegen.NewTypeGenerator(mr, encMeta, "github.com/aphoh/go-substrate-gen/palletgen")
		palletGen := NewPalletGenerator(&pal, &tg)

		res, isSome, err := palletGen.GenerateStorage("github.com/aphoh/go-substrate-gen/palletgen")
		if isSome {
			require.False(t, strings.Contains(res, "%!v(PANIC="), "Generated code contains errors in pallet", string(pal.Name))
		}
		require.NoError(t, err)

		res, isSome, err = palletGen.GenerateCalls("github.com/aphoh/go-substrate-gen/palletgen")
		require.NoError(t, err)
		if isSome {
			require.False(t, strings.Contains(res, "%!v(PANIC="), "Generated code contains errors in pallet", string(pal.Name))
		}
	}
}

// Enable this to see some sample output
func noTestSamplePalletOutput(t *testing.T) {
	inp, err := ioutil.ReadFile("../polka-meta.json")
	require.NoError(t, err)
	mr, encMeta, err := metadata.ParseMetadata(inp)

	tg := typegen.NewTypeGenerator(mr, encMeta, "github.com/aphoh/go-substrate-gen/palletgen")
	palletGen := NewPalletGenerator(&mr.Pallets[6], &tg)

	storage, isSome, err := palletGen.GenerateStorage("github.com/aphoh/go-substrate-gen/palletgen")
	require.NoError(t, err)
	if isSome {
		ioutil.WriteFile("test_storage.go", []byte(storage), 0644)
	}
	calls, isSome, err := palletGen.GenerateCalls("github.com/aphoh/go-substrate-gen/palletgen")
	require.NoError(t, err)
	if isSome {
		ioutil.WriteFile("test_calls.go", []byte(calls), 0644)
	}
	types := tg.GetGenerated()

	ioutil.WriteFile("test_types.go", []byte(types), 0644)

	t.Fail()
}
