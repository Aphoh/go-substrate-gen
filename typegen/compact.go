package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func (tg *TypeGenerator) GenCompact(v *types.Si1TypeDefCompact, mt *types.PortableTypeV14) (GeneratedType, error) {
	innerT, err := tg.GetType(v.Type.Int64())
	if err != nil {
		return nil, err
	}

	var g GeneratedType
	if eg, ok := innerT.(*Gend); ok {
		if eg.Name != "U128" {
			return nil, fmt.Errorf("Unsupported compact type %v", v)
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
			return nil, fmt.Errorf("Unsupported compact type %v", v)
		}
	} else {
		return nil, fmt.Errorf("Unsupported compact type %v", v)
	}

	tg.generated[mt.ID.Int64()] = g
	return g, nil
}
