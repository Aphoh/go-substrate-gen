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

func (rg *PalletGenerator) GenerateStorage(pkgFilePath string) (string, error) {
	sgen := storagegen.NewStorageGenerator(pkgFilePath, &rg.pallet.Storage, rg.tygen)
	err := sgen.Generate()
	if err != nil {
		return "", err
	}
	// Do something with sgen.F
	return fmt.Sprintf("%#v", sgen.F), nil
}

func (rg *PalletGenerator) GenerateCalls(pkgFilePath string) (string, error) {
	callGen := callgen.NewCallGenerator(pkgFilePath, rg.pallet, rg.tygen)
	err := callGen.Generate()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%#v", callGen.F), nil
}
