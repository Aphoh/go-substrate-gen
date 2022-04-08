package palletgen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/pal"
	"github.com/aphoh/go-substrate-gen/palletgen/callgen"
	"github.com/aphoh/go-substrate-gen/palletgen/storagegen"
	"github.com/aphoh/go-substrate-gen/typegen"
)

type PalletGenerator struct {
	pallet *pal.Pallet
	tygen  *typegen.TypeGenerator
}

func NewPalletGenerator(pallet *pal.Pallet, tygen *typegen.TypeGenerator) PalletGenerator {
	return PalletGenerator{pallet: pallet, tygen: tygen}
}

func (rg *PalletGenerator) GenerateStorage(pkgFilePath string) (string, bool, error) {
  if rg.pallet.Storage == nil {
    return "", false, nil
  }
	sgen := storagegen.NewStorageGenerator(pkgFilePath, rg.pallet.Storage, rg.tygen)
	err := sgen.Generate()
	if err != nil {
		return "", false, err
	}
	// Do something with sgen.F
	return fmt.Sprintf("%#v", sgen.F), true, nil
}

func (rg *PalletGenerator) GenerateCalls(pkgFilePath string) (string, bool, error) {
  if rg.pallet.Calls == nil {
    return "", false, nil
  }
	callGen := callgen.NewCallGenerator(pkgFilePath, rg.pallet, rg.tygen)
	err := callGen.Generate()
	if err != nil {
		return "", false, err
	}

	return fmt.Sprintf("%#v", callGen.F), true, nil
}
