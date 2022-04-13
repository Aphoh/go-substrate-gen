package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

// These are really just vecs with a given significance ordering. We'll ignore that for now and just hand back a slice.
func (tg *TypeGenerator) GenBitsequence(bs *tdk.TDBitsequence, mt *tdk.MType) (GeneratedType, error) {
	innerTy, err := tg.GetType(bs.BitStoreTypeId)
	if err != nil {
		return nil, err
	}
	switch v := innerTy.(type) {
	case *PrimitiveGend:
		g := SliceGend{
			Inner: innerTy,
			MTy:   mt,
		}
		tg.generated[mt.Id] = &g
		return &g, nil

	default:
		return nil, fmt.Errorf("Bitsequence with nonprimitive type %v, typeid=%v", v.DisplayName(), mt.Id)
	}
}
