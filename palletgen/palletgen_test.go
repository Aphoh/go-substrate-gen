package palletgen

import (
	"fmt"
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

	tg := typegen.NewTypeGenerator(&mr, "palletgen")
	palletGen := NewPalletGenerator(mr.Pallets, &tg, "github.com/aphoh/go-substrate-gen/palletgen")

	psystem, err := palletGen.GeneratePallet(0, "palletgen")
	require.NoError(t, err)
	tgen := fmt.Sprintf("%#v", tg.F)

	ioutil.WriteFile("psystem.go", []byte(psystem), 0644)

	ioutil.WriteFile("types_out.go", []byte(tgen), 0644)

	t.Fail()
}
