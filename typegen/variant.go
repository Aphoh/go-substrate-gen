package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/dave/jennifer/jen"
)

const SCALE = "github.com/centrifuge/go-substrate-rpc-client/v4/scale"

func (tg *TypeGenerator) GenVariant(v *tdk.TDVariant, mt *tdk.MType) (*gend, error) {
	if len(v.Variants) == 0 {
		g := gend{
			name: "struct{}",
			id:   mt.Id,
		}
		tg.generated[mt.Id] = g
		return &g, nil
	}

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
				errorCheckG(g2)
				for j := range variant.Fields {
					g2.Id("err").Op("=").Id("encoder").Dot("Encode").Call(jen.Id("ty").Dot(variantFieldNames[i][j]))
					errorCheckG(g2)
				}
				// return ok
				g2.Return(jen.Nil())
			})
		}
		// Didn't hit any variant, return error
		g1.Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("Unrecognized variant")))
	})

	// func (ty *g.name) Decode(decoder scale.Decoder) (err error) {...}
	tg.f.Func().Params(
		jen.Id("ty").Op("*").Id(g.name)).Id("Decode").Params(jen.Id("decoder").Qual(SCALE, "Decoder")).Params(
		jen.Err().Error(),
	).BlockFunc(func(g1 *jen.Group) {
		// variant, err := decoder.ReadOneByte()
		g1.Id("variant, err").Op(":=").Id("decoder").Dot("ReadOneByte").Call()
		errorCheckG(g1)
		// switch variant {..}
		g1.Switch(jen.Id("variant")).BlockFunc(func(g2 *jen.Group) {
			for i, variant := range v.Variants {
				// case i:
				g2.Case(jen.Lit(i)).BlockFunc(func(g3 *jen.Group) {
					// ty.isVariantI = true
					g3.Id("ty").Dot(variantIsNames[i]).Op("=").True()
					// decode remaining fields
					for j := range variant.Fields {
						g3.Err().Op("=").Id("decoder").Dot("Decode").Call(
							jen.Op("&").Id("ty").Dot(variantFieldNames[i][j]),
						)
						errorCheckG(g3)
					}
					g3.Return()
				})
			}
			g2.Default().Block(jen.Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("Unrecognized variant"))))
		})
	})

	return g, nil
}

func errorCheckG(s *jen.Group) {
	s.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.Err()),
	)
}
