package palletgen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/palletgen/callgen"
	"github.com/aphoh/go-substrate-gen/palletgen/storagegen"
	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// The palletgenerator is responsible for determing if a pallet needs storage or calls generated,
// and generating them if necessary with a StorageGenerator or CallGenerator.
type PalletGenerator struct {
	pallet *types.PalletMetadataV14
	tygen  *typegen.TypeGenerator
}

func NewPalletGenerator(pallet *types.PalletMetadataV14, tygen *typegen.TypeGenerator) PalletGenerator {
	return PalletGenerator{pallet: pallet, tygen: tygen}
}

// Generate all storage calls for the pallet, and return the file as a string
func (rg *PalletGenerator) GenerateStorage(pkgFilePath string) (string, bool, error) {
	if !rg.pallet.HasStorage {
		return "", false, nil
	}
	sgen := storagegen.NewStorageGenerator(pkgFilePath, &rg.pallet.Storage, rg.tygen)
	err := sgen.Generate()
	if err != nil {
		return "", false, err
	}
	// Do something with sgen.F
	return fmt.Sprintf("%#v", sgen.F), true, nil
}

// Generate all calls (extrinsics) for the pallet, and return the file as a string
func (rg *PalletGenerator) GenerateCalls(pkgFilePath string) (string, bool, error) {
	if !rg.pallet.HasCalls {
		return "", false, nil
	}
	callGen := callgen.NewCallGenerator(pkgFilePath, rg.pallet, rg.tygen)
	err := callGen.Generate()
	if err != nil {
		return "", false, err
	}

	return fmt.Sprintf("%#v", callGen.F), true, nil
}
