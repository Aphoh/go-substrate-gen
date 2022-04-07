package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/utils"
)


func (tg *TypeGenerator) GenCompact(v *tdk.TDCompact, mt *tdk.MType) (GeneratedType, error) {
	innerT, err := tg.GetType(v.TypeId)
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
			}
		default:
			return nil, fmt.Errorf("Unsupported compact type %v", v)
		}
	} else {
		return nil, fmt.Errorf("Unsupported compact type %v", v)
	}

	tg.generated[mt.Id] = g
	return g, nil
}
