package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/utils"
)

func (tg *TypeGenerator) GenPrimitive(primitive *tdk.TDPrimitive, mt *tdk.MType) (GeneratedType, error) {
	var g GeneratedType
	switch *primitive {
	case "U8":
		g = &PrimitiveGend{PrimName: "byte", MTy: mt}
	case "U16":
		g = &PrimitiveGend{PrimName: "uint16", MTy: mt}
	case "U32":
		g = &PrimitiveGend{PrimName: "uint32", MTy: mt}
	case "U64":
		g = &PrimitiveGend{PrimName: "uint64", MTy: mt}
	case "U128":
		g = &Gend{
			Name: "U128",
			Pkg:  utils.CTYPES,
			MTy:  mt,
		}
	case "Str":
		g = &PrimitiveGend{PrimName: "string", MTy: mt}
	case "Bool":
		g = &PrimitiveGend{PrimName: "bool", MTy: mt}
	default:
		return nil, fmt.Errorf("Unsupported primitive %s", string(*primitive))
	}
	tg.generated[mt.Id] = g
	return g, nil
}
