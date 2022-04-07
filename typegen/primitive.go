package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

func (tg *TypeGenerator) GenPrimitive(primitive *tdk.TDPrimitive, id string) (*Gend, error) {
	g := Gend{Id: id}
	switch *primitive {
	case "U8":
		g.Name = "byte"
	case "U16":
		g.Name = "uint16"
	case "U32":
		g.Name = "uint32"
	case "U64":
		g.Name = "uint64"
	case "U128":
		g.Name = "ctypes.U128"
	case "Str":
		g.Name = "string"
	case "Bool":
		g.Name = "bool"
	default:
		return nil, fmt.Errorf("Unsupported primitive %s", string(*primitive))
	}
  tg.generated[id] = g
	return &g, nil
}
