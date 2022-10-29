package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// Generate and return a SCALE-encoded compact.
// A compact type is a compacted unsigned integer, such that an integer between 1 and 2^{i-1} only takes i+2 bits.
// The leading two bits describe the rest of the encoding.
// See https://docs.substrate.io/reference/scale-codec/#fn-1 for a description of the encoding
func (tg *TypeGenerator) GenCompact(v *types.Si1TypeDefCompact, mt *types.PortableTypeV14) (GeneratedType, error) {
	innerT, err := tg.GetType(v.Type.Int64())
	if err != nil {
		return nil, err
	}

	var g GeneratedType
	if eg, ok := innerT.(*Gend); ok {
		// This check for if the inner type is a defined type is only necessary because go does not
		// have uint128 as a primitive. If it did, we would only look for primitive types inside a
		// compact
		if eg.Name != "U128" {
			return nil, fmt.Errorf("unsupported compact type %v", v)
		}
		g = &Gend{
			Name: "UCompact",
			Pkg:  utils.CTYPES,
			MTy:  mt,
		}
	} else if sg, ok := innerT.(*PrimitiveGend); ok {
		switch sg.PrimName {
		case "struct{}":
			// Just use the same struct
			// Realistically, this should only ever be a big int or similar
			g = sg
		case "uint16":
			fallthrough
		case "uint32":
			fallthrough
		case "uint64":
			g = &Gend{
				Name: "UCompact",
				Pkg:  utils.CTYPES,
				MTy:  mt,
			}
		default:
			return nil, fmt.Errorf("unsupported compact type %v", v)
		}
	} else {
		return nil, fmt.Errorf("unsupported compact type %v", v)
	}

	tg.generated[mt.ID.Int64()] = g
	return g, nil
}
