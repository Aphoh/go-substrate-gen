package palletgen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/pal"
	"github.com/aphoh/go-substrate-gen/palletgen/storagegen"
	"github.com/aphoh/go-substrate-gen/typegen"
)

type PalletGenerator struct {
	pallets   []pal.Pallet
	tygen     *typegen.TypeGenerator
	typesPath string
}

func NewPalletGenerator(pallets []pal.Pallet, tygen *typegen.TypeGenerator, typesPath string) PalletGenerator {
	return PalletGenerator{pallets: pallets, tygen: tygen, typesPath: typesPath}
}

func (rg *PalletGenerator) GeneratePallet(i uint32, palletPkgName string) (string, error) {
	p := rg.pallets[i]
	sgen := storagegen.NewStorageGenerator(palletPkgName, &p.Storage, rg.tygen, rg.typesPath)
	err := sgen.Generate()
	if err != nil {
		return "", err
	}
	// Do something with sgen.F
	return fmt.Sprintf("%#v", sgen.F), nil
}
