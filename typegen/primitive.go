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
    g.Global = true
	case "U16":
		g.Name = "uint16"
    g.Global = true
	case "U32":
		g.Name = "uint32"
    g.Global = true
	case "U64":
		g.Name = "uint64"
    g.Global = true
	case "U128":
		g.Name = "types.U128"
	case "Str":
		g.Name = "string"
    g.Global = true
	case "Bool":
		g.Name = "bool"
    g.Global = true
	default:
		return nil, fmt.Errorf("Unsupported primitive %s", string(*primitive))
	}
  tg.generated[id] = g
	return &g, nil
}
