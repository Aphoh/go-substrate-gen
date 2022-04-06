package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/tdk"
)

func (tg *TypeGenerator) GenPrimitive(primitive *tdk.TDPrimitive, id string) (*gend, error) {
	g := gend{id: id}
	switch *primitive {
	case "U8":
		g.name = "byte"
	case "U16":
		g.name = "uint16"
	case "U32":
		g.name = "uint32"
	case "U64":
		g.name = "uint64"
	case "U128":
		g.name = "ctypes.U128"
	case "Str":
		g.name = "string"
	case "Bool":
		g.name = "bool"
	default:
		return nil, fmt.Errorf("Unsupported primitive %s", string(*primitive))
	}
  tg.generated[id] = g
	return &g, nil
}
