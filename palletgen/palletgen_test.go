package palletgen

import (
	"io/ioutil"
	"testing"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/stretchr/testify/require"
)

func TestGenBigMetadata(t *testing.T) {
	inp, err := ioutil.ReadFile("../json-gen/meta.json")
	require.NoError(t, err)
	mr, err := metadata.ParseMetadata(inp)

	tg := typegen.NewTypeGenerator(&mr, "github.com/aphoh/go-substrate-gen/palletgen")
	palletGen := NewPalletGenerator(&mr.Pallets[18], &tg)

	storage, err := palletGen.GenerateStorage("github.com/aphoh/go-substrate-gen/palletgen")
	calls, err := palletGen.GenerateCalls("github.com/aphoh/go-substrate-gen/palletgen")
	require.NoError(t, err)
	types := tg.GetGenerated()

	ioutil.WriteFile("test_storage.go", []byte(storage), 0644)
	ioutil.WriteFile("test_calls.go", []byte(calls), 0644)
	ioutil.WriteFile("test_types.go", []byte(types), 0644)

	t.Fail()
}
