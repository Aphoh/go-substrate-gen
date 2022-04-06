package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/tdk"
	"github.com/dave/jennifer/jen"
)

func (tg *TypeGenerator) GenTuple(tup *tdk.TDTuple, mt *tdk.MType) (*gend, error) {
	if len(*tup) == 0 {
		g := gend{
			name: "struct{}",
			id:   mt.Id,
		}
		tg.generated[mt.Id] = g
		return &g, nil
	}

	tn := asName("Tuple", mt.Id)

	g := gend{
		name: tn,
		id:   mt.Id,
	}

	tg.generated[mt.Id] = g
	code := []jen.Code{}
	for i, te := range *tup {
		ty, err := tg.GetType(te)
		if err != nil {
			return nil, err
		}
		code = append(code, jen.Id(asName("Elem", fmt.Sprint(i))).Id(ty.name))
	}
	tg.f.Comment(fmt.Sprintf("Tuple type %v", mt.Id))
	tg.f.Type().Id(tn).Struct(code...)

	return &g, nil
}
