package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/dave/jennifer/jen"
)

func (tg *TypeGenerator) GenTuple(tup *tdk.TDTuple, mt *tdk.MType) (*Gend, error) {
	if len(*tup) == 0 {
		g := Gend{
			Name: "struct{}",
			Id:   mt.Id,
		}
		tg.generated[mt.Id] = g
		return &g, nil
	}

	tn := utils.AsName("Tuple", mt.Id)

	g := Gend{
		Name: tn,
		Id:   mt.Id,
	}

	tg.generated[mt.Id] = g
	code := []jen.Code{}
	for i, te := range *tup {
		ty, err := tg.GetType(te)
		if err != nil {
			return nil, err
		}
		code = append(code, jen.Id(utils.AsName("Elem", fmt.Sprint(i))).Id(ty.Name))
	}
	tg.f.Comment(fmt.Sprintf("Tuple type %v", mt.Id))
	tg.f.Type().Id(tn).Struct(code...)

	return &g, nil
}
