package palletgen

import (
	"path"

	"github.com/aphoh/go-substrate-gen/metadata/pal"
	"github.com/aphoh/go-substrate-gen/palletgen/storagegen"
	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/dave/jennifer/jen"
)

type PalletGenerator struct {
	f         *jen.File
	pallets   []pal.Pallet
	tygen     *typegen.TypeGenerator
	pkgPath   string
	typesPath string
}

func NewPalletGenerator(pkgPath string, pallets []pal.Pallet, tygen *typegen.TypeGenerator, typesPath string) PalletGenerator {
	f := jen.NewFilePath(pkgPath)
	return PalletGenerator{f: f, pallets: pallets, tygen: tygen, typesPath: typesPath}
}

func (rg *PalletGenerator) GenerateStorage(i uint32) error {
	p := rg.pallets[i]
	spath := path.Join(rg.pkgPath, "/storage")
	sgen := storagegen.NewStorageGenerator(spath, &p.Storage, rg.tygen, rg.typesPath)
	err := sgen.Generate()
	if err != nil {
		return err
	}
  // Do something with sgen.F
  return nil
}
