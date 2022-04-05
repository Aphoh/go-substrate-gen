package gen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/tdk"
	"github.com/dave/jennifer/jen"
)

const SCALE = "github.com/centrifuge/go-substrate-rpc-client/v4/scale"

func (tg *TypeGenerator) GenVariant(v *tdk.TDVariant, mt *tdk.MType) (*gend, error) {
	g, err := tg.getStructName(mt)
	if err != nil {
		return nil, err
	}

	inner := []jen.Code{}

	variantIsNames := []string{}
	variantFieldNames := [][]string{}

	for i, variant := range v.Variants {
		vIsName := asName("Is", variant.Name)
		variantIsNames = append(variantIsNames, vIsName)
		inner = append(inner, jen.Id(vIsName).Bool())
		variantFieldNames = append(variantFieldNames, []string{})

		for j, f := range variant.Fields {
			fc, fieldName, err := tg.fieldCode(f, "As_"+variant.Name, fmt.Sprint(j))
			if err != nil {
				return nil, err
			}
			variantFieldNames[i] = append(variantFieldNames[i], fieldName)
			inner = append(inner, fc...)
		}

	}

	tg.f.Comment(fmt.Sprintf("Generated %v with id=%v", asName(mt.Ty.Path...), mt.Id))
	tg.f.Type().Id(g.name).Struct(inner...)

	// func (ty *g.name) Encode(encoder scale.Encoder) (err error) {...}
	tg.f.Func().Params(
		jen.Id("ty").Op("*").Id(g.name)).Id("Encode").Params(jen.Id("encoder").Qual(SCALE, "Encoder")).Params(
		jen.Err().Error(),
	).BlockFunc(func(g1 *jen.Group) {
		// for each variant, check if variant
		for i, variant := range v.Variants {
			g1.If(jen.Id("ty").Dot(variantIsNames[i])).BlockFunc(func(g2 *jen.Group) {
				// if is variant, encode stuff for variant
				g2.Err().Op("=").Id("encoder").Dot("PushByte").Call(jen.Lit(i))
				g2.If(jen.Err().Op("!=").Nil()).Block(
					jen.Return(jen.Err()),
				)
				for j := range variant.Fields {
					g2.Id("err").Op("=").Id("encoder").Dot("Encode").Call(jen.Id("ty").Dot(variantFieldNames[i][j]))
					g2.If(jen.Err().Op("!=").Nil()).Block(
						jen.Return(jen.Err()),
					)
				}
				// return ok
				g2.Return(jen.Nil())
			})
		}
		// Didn't hit any variant, return error
		g1.Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("Unrecognized variant")))
	})

	return g, nil
}
