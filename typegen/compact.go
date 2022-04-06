package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

func (tg *TypeGenerator) GenCompact(v *tdk.TDCompact, mt *tdk.MType) (*gend, error) {
	innerT, err := tg.GetType(v.TypeId)
	if err != nil {
		return nil, err
	}

	var name string
	switch innerT.name {
	case "struct{}":
		name = innerT.name
	case "uint32":
		fallthrough
	case "uint64":
		fallthrough
	case "ctypes.U128":
		name = "ctypes.UCompact"
	default:
		panic(fmt.Sprintf("Unknown compact type %v", innerT.name))
	}
	g := gend{
		name: name,
		id:   mt.Id,
	}
	tg.generated[mt.Id] = g
	return &g, nil
}
