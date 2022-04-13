package typegen

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// These are really just vecs with a given significance ordering. We'll ignore that for now and just hand back a slice.
func (tg *TypeGenerator) GenBitsequence(bs *types.Si1TypeDefBitSequence, mt *types.PortableTypeV14) (GeneratedType, error) {
	innerTy, err := tg.GetType(bs.BitStoreType.Int64())
	if err != nil {
		return nil, err
	}
	switch v := innerTy.(type) {
	case *PrimitiveGend:
		g := SliceGend{
			Inner: innerTy,
			MTy:   mt,
		}
		tg.generated[mt.ID.Int64()] = &g
		return &g, nil

	default:
		return nil, fmt.Errorf("Bitsequence with nonprimitive type %v, typeid=%v", v.DisplayName(), mt.ID.Int64())
	}
}
