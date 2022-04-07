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
		g = &PrimitiveGend{PrimName: "struct{}"}
		tg.generated[mt.Id] = g
		return g, nil
	}

	tn := utils.AsName("Tuple", mt.Id)
	g = &Gend{
		Name: tn,
		Pkg:  tg.pkgPath,
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

	// func (tup *Tuple__) TupleEncodeEach() ([][]byte, error)
	tg.F.Func().Parens(
		jen.Id("tup").Op("*").Id(tn),
	).Id(utils.TupleEncodeEach).Call().Parens(
		jen.List(jen.Index().Index().Byte(), jen.Error()),
	).BlockFunc(func(g *jen.Group) {
		g.Id("ba").Op(":=").Index().Index().Byte().Values()
		g.Var().Id("bytes").Index().Byte()
		g.Var().Err().Error()
		for _, f := range fieldNames {
			// bytes, err := ctypes.EncodeToBytes(tup.field)
			g.List(jen.Id("bytes"), jen.Err()).Op("=").Qual(utils.CTYPES, "EncodeToBytes").Call(
				jen.Id("tup").Dot(f),
			)
			utils.ErrorCheckWithNil(g)
			// ba = append(ba, bytes)
			g.Id("ba").Op("=").Append(jen.List(jen.Id("ba"), jen.Id("bytes")))
		}
		g.Return(jen.List(jen.Id("ba"), jen.Nil()))
	})

	return g, nil
}
