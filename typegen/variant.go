package typegen

import (
	"fmt"
	"strconv"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/dave/jennifer/jen"
)

const SCALE = "github.com/centrifuge/go-substrate-rpc-client/v4/scale"

func (tg *TypeGenerator) GenVariant(v *tdk.TDVariant, mt *tdk.MType) (GeneratedType, error) {
	if len(v.Variants) == 0 {
		g := &PrimitiveGend{
			PrimName: "struct{}",
			MTy:      mt,
		}
		tg.generated[mt.Id] = g
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
	tg.generated[mt.Id] = g

	inner := []jen.Code{}

	variantIsNames := []string{}
	variantFieldNames := [][]string{}

	for i, variant := range v.Variants {
		vIsName := utils.AsName("Is", variant.Name)
		variantIsNames = append(variantIsNames, vIsName)
		inner = append(inner, jen.Id(vIsName).Bool())
		variantFieldNames = append(variantFieldNames, []string{})

		for j, f := range variant.Fields {
			fc, fieldName, err := tg.fieldCode(f, utils.AsName("As", variant.Name), fmt.Sprint(j))
			if err != nil {
				return nil, err
			}
			variantFieldNames[i] = append(variantFieldNames[i], fieldName)
			inner = append(inner, fc...)
		}

	}

	tg.F.Comment(fmt.Sprintf("Generated %v with id=%v", utils.AsName(mt.Ty.Path...), mt.Id))
	tg.F.Type().Id(g.Name).Struct(inner...)

	// func (ty *g.name) Encode(encoder scale.Encoder) (err error) {...}
	tg.F.Func().Params(
		jen.Id("ty").Op("*").Id(g.Name)).Id("Encode").Params(jen.Id("encoder").Qual(SCALE, "Encoder")).Params(
		jen.Err().Error(),
	).BlockFunc(func(g1 *jen.Group) {
		// for each variant, check if variant
		for i, variant := range v.Variants {
      // This index is not necessarily the index that it appears at in the list
      varI, err := strconv.Atoi(variant.Index)
      if err != nil {
        panic(fmt.Sprintf("Invalid index given in variant %v", variant.Index))
      }
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
		jen.Id("ty").Op("*").Id(g.Name)).Id("Decode").Params(jen.Id("decoder").Qual(SCALE, "Decoder")).Params(
		jen.Err().Error(),
	).BlockFunc(func(g1 *jen.Group) {
		// variant, err := decoder.ReadOneByte()
		g1.Id("variant, err").Op(":=").Id("decoder").Dot("ReadOneByte").Call()
		utils.ErrorCheckG(g1)
		// switch variant {..}
		g1.Switch(jen.Id("variant")).BlockFunc(func(g2 *jen.Group) {
			for i, variant := range v.Variants {
        // This index is not necessarily the index that it appears at in the list
				varI, err := strconv.Atoi(variant.Index)
				if err != nil {
					panic(fmt.Sprintf("Invalid index given in variant %v", variant.Index))
				}
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

	return g, nil
}
