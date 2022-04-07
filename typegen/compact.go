package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

func (tg *TypeGenerator) GenCompact(v *tdk.TDCompact, mt *tdk.MType) (*Gend, error) {
	innerT, err := tg.GetType(v.TypeId)
	if err != nil {
		return nil, err
	}

	var name string
	switch innerT.Name {
	case "struct{}":
		name = innerT.Name
	case "uint32":
		fallthrough
	case "uint64":
		fallthrough
	case "ctypes.U128":
		name = "ctypes.UCompact"
	default:
		panic(fmt.Sprintf("Unknown compact type %v", innerT.Name))
	}
	g := Gend{
		Name: name,
		Id:   mt.Id,
	}
	tg.generated[mt.Id] = g
	return &g, nil
}
