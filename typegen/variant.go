package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

const SCALE = "github.com/centrifuge/go-substrate-rpc-client/v4/scale"

func (tg *TypeGenerator) GenVariant(v *types.Si1TypeDefVariant, mt *types.PortableTypeV14) (GeneratedType, error) {
	if len(v.Variants) == 0 {
		g := &PrimitiveGend{
			PrimName: "struct{}",
			MTy:      mt,
		}
		tg.generated[mt.ID.Int64()] = g
		return g, nil
	}

	sName, err := tg.getStructName(mt)
	if err != nil {
		return nil, err
	}
	g := &Gend{
		Name: sName,
		Pkg:  tg.PkgPath,
		MTy:  mt,
	}
	tg.generated[mt.ID.Int64()] = g

	inner := []jen.Code{jen.Id("variant").Id("uint8")}

	variantIsNames := []string{}
	variantFieldNames := [][]string{}

	for i, variant := range v.Variants {
		vIsName := utils.AsName("Is", string(variant.Name))
		variantIsNames = append(variantIsNames, vIsName)
		inner = append(inner, jen.Id(vIsName).Bool())
		variantFieldNames = append(variantFieldNames, []string{})

		for j, f := range variant.Fields {
			useTypeName := len(variant.Fields) > 1
			fc, fieldName, err := tg.fieldCode(f, utils.AsName("As", string(variant.Name)), fmt.Sprint(j), useTypeName)
			if err != nil {
				return nil, err
			}
			variantFieldNames[i] = append(variantFieldNames[i], fieldName)
			inner = append(inner, fc...)
		}

	}

	tg.F.Comment(fmt.Sprintf("Generated %v with id=%v", utils.AsName(utils.PathStrs(mt.Type.Path)...), mt.ID.Int64()))
	tg.F.Type().Id(g.Name).Struct(inner...)

	// IMPORTANT: we only implement encode for the actual type, not the pointer (ty *g.name),
	// because otherwise reflection will fail to see that the object has the method if it's embedded
	// in another type
	// func (ty g.name) Encode(encoder scale.Encoder) (err error) {...}
	tg.F.Func().Params(
		jen.Id("ty").Id(g.Name),
	).Id("Encode").Params(jen.Id("encoder").Qual(SCALE, "Encoder")).Params(
		jen.Err().Error(),
	).BlockFunc(func(g1 *jen.Group) {
		// for each variant, check if variant
		for i, variant := range v.Variants {
			// This index is not necessarily the index that it appears at in the list
			varI := uint8(variant.Index)
			g1.If(jen.Id("ty").Dot(variantIsNames[i])).BlockFunc(func(g2 *jen.Group) {
				// if is variant, encode stuff for variant
				g2.Err().Op("=").Id("encoder").Dot("PushByte").Call(jen.Lit(varI))
				utils.ErrorCheckG(g2)
				for j := range variant.Fields {
					g2.Id("err").Op("=").Id("encoder").Dot("Encode").Call(jen.Id("ty").Dot(variantFieldNames[i][j]))
					utils.ErrorCheckG(g2)
				}
				// return ok
				g2.Return(jen.Nil())
			})
		}
		// Didn't hit any variant, return error
		g1.Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("Unrecognized variant")))
	})

	// func (ty *g.name) Decode(decoder scale.Decoder) (err error) {...}
	tg.F.Func().Params(
		jen.Id("ty").Op("*").Id(g.Name),
	).Id("Decode").Params(jen.Id("decoder").Qual(SCALE, "Decoder")).Params(
		jen.Err().Error(),
	).BlockFunc(func(g1 *jen.Group) {
		// variant, err := decoder.ReadOneByte()
		g1.List(jen.Id("ty").Dot("variant"), jen.Err()).Op("=").Id("decoder").Dot("ReadOneByte").Call()
		utils.ErrorCheckG(g1)
		// switch variant {..}
		g1.Switch(jen.Id("ty").Dot("variant")).BlockFunc(func(g2 *jen.Group) {
			for i, variant := range v.Variants {
				// This index is not necessarily the index that it appears at in the list
				varI := uint8(variant.Index)
				g2.Case(jen.Lit(varI)).BlockFunc(func(g3 *jen.Group) {
					// ty.isVariantI = true
					g3.Id("ty").Dot(variantIsNames[i]).Op("=").True()
					// decode remaining fields
					for j := range variant.Fields {
						g3.Err().Op("=").Id("decoder").Dot("Decode").Call(
							jen.Op("&").Id("ty").Dot(variantFieldNames[i][j]),
						)
						utils.ErrorCheckG(g3)
					}
					g3.Return()
				})
			}
			g2.Default().Block(jen.Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("Unrecognized variant"))))
		})
	})

	tg.F.Func().Params(
		jen.Id("ty").Op("*").Id(g.Name),
	).Id("Variant").Call().Id("uint8").Block(
		jen.Return().Id("ty").Dot("variant"),
	)

	return g, nil
}
