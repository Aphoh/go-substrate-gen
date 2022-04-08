package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/dave/jennifer/jen"
)

func (tg *TypeGenerator) GenTuple(tup *tdk.TDTuple, mt *tdk.MType) (GeneratedType, error) {
	var g GeneratedType
	// Empty tuple maps to struct{}
	if len(*tup) == 0 {
		g = &PrimitiveGend{PrimName: "struct{}", MTy: mt}
		tg.generated[mt.Id] = g
		return g, nil
	}

	tn := utils.AsName("Tuple", mt.Id)
	g = &Gend{
		Name: tn,
		Pkg:  tg.PkgPath,
		MTy:  mt,
	}

	tg.generated[mt.Id] = g

	code := []jen.Code{}
	fieldNames := []string{}
	for i, te := range *tup {
		ty, err := tg.GetType(te)
		if err != nil {
			return nil, err
		}
		fName := utils.AsName("Elem", fmt.Sprint(i))
		fieldNames = append(fieldNames, fName)
		code = append(code, jen.Id(fName).Custom(utils.TypeOpts, ty.Code()))
	}
	tg.F.Comment(fmt.Sprintf("Tuple type %v", mt.Id))
	tg.F.Type().Id(tn).Struct(code...)
	return g, nil
}
