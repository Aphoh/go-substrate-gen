package typegen

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// Generate and return a go array of type `arr.Type` from the metadata array `arr` and the metadata
// type `mt`
func (tg *TypeGenerator) GenArray(arr *types.Si1TypeDefArray, mt *types.PortableTypeV14) (GeneratedType, error) {
	tyGend, err := tg.GetType(arr.Type.Int64())
	if err != nil {
		return nil, err
	}
	g := ArrayGend{
		Inner: tyGend,
		Len:   int(arr.Len),
		MTy:   mt,
	}
	tg.generated[mt.ID.Int64()] = &g
	return &g, nil
}
