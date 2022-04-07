package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/utils"
)

func (tg *TypeGenerator) GenPrimitive(primitive *tdk.TDPrimitive, id string) (GeneratedType, error) {
  var g GeneratedType
	switch *primitive {
	case "U8":
		g = &PrimitiveGend{PrimName: "byte"}
	case "U16":
		g = &PrimitiveGend{PrimName: "uint16"}
	case "U32":
		g = &PrimitiveGend{PrimName: "uint32"}
	case "U64":
		g = &PrimitiveGend{PrimName: "uint64"}
	case "U128":
		g = &Gend{
			Name: "U128",
			Pkg:  utils.CTYPES,
		}
	case "Str":
		g = &PrimitiveGend{PrimName: "string"}
	case "Bool":
		g = &PrimitiveGend{PrimName: "bool"}
	default:
		return nil, fmt.Errorf("Unsupported primitive %s", string(*primitive))
	}
	tg.generated[id] = g
	return g, nil
}
