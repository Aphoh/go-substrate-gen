package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

func (tg *TypeGenerator) GenTuple(tup *types.Si1TypeDefTuple, mt *types.PortableTypeV14) (GeneratedType, error) {
	var g GeneratedType
	// Empty tuple maps to struct{}
	if len(*tup) == 0 {
		g = &PrimitiveGend{PrimName: "struct{}", MTy: mt}
		tg.generated[mt.ID.Int64()] = g
		return g, nil
	} else if len(*tup) == 1 {
		g, err := tg.GetType((*tup)[0].Int64())
		if err != nil {
			return nil, err
		}
		tg.generated[mt.ID.Int64()] = g
		return g, nil
	}

	tn := utils.AsName("Tuple", fmt.Sprint(mt.ID.Int64()))
	g = &Gend{
		Name: tn,
		Pkg:  tg.PkgPath,
		MTy:  mt,
	}

	tg.generated[mt.ID.Int64()] = g

	code := []jen.Code{}
	fieldNames := []string{}
	for i, te := range *tup {
		ty, err := tg.GetType(te.Int64())
		if err != nil {
			return nil, err
		}
		fName := utils.AsName("Elem", fmt.Sprint(i))
		fieldNames = append(fieldNames, fName)
		code = append(code, jen.Id(fName).Custom(utils.TypeOpts, ty.Code()))
	}
	tg.F.Comment(fmt.Sprintf("Tuple type %v", mt.ID.Int64()))
	tg.F.Type().Id(tn).Struct(code...)
	return g, nil
}
