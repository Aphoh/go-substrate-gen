package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

const SCALE = "github.com/centrifuge/go-substrate-rpc-client/v4/scale"

// Generate and return a go struct which represents a rust variant. When generated, this will also
// define the struct in `types/types.go`, as well as define an `Encode`, `Decode`, and `Variant`
// method on it.
//
// example variant:
//
//	type Result struct {
//		IsOk        bool
//		AsOkField0  struct{}
//		IsErr       bool
//		AsErrField0 DispatchError
//	}
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
	numVariants := len(v.Variants)
	vGend := &VariantGend{
		Gend: Gend{
			Name: sName,
			Pkg:  tg.PkgPath,
			MTy:  mt,
		},
		IsVarFields: make([]GenField, numVariants),
		AsVarFields: make([][]GenField, numVariants),
		Indices:     make([]uint8, numVariants),
	}
	tg.generated[mt.ID.Int64()] = vGend

	inner := []jen.Code{}

	for i, variant := range v.Variants {
		vIsName := utils.AsName("Is", string(variant.Name))
		inner = append(inner, jen.Id(vIsName).Bool())

		vGend.IsVarFields[i] = GenField{Name: vIsName, IsPtr: false}
		vGend.Indices[i] = uint8(variant.Index)

		usePtr, _, _ := tg.variantFieldUsesPointer(variant)

		// There may be many fields in a variant, as rust enums can have values that are full on structs
		// In that case, instead of the struct going: isVarA, varAField, isVarB ..., it instead goes
		// isVarA, varAField0, varAField1, .., varAFieldn, isVarB, etc
		for j, f := range variant.Fields {
			useTypeName := len(variant.Fields) > 1
			gf, err := tg.fieldCode(f, utils.AsName("As", string(variant.Name)), fmt.Sprint(j), useTypeName, usePtr)
			if err != nil {
				return nil, err
			}
			vGend.AsVarFields[i] = append(vGend.AsVarFields[i], *gf)
			inner = append(inner, gf.Code...)
		}

	}

	// Generate the variant type itself
	tg.F.Comment(fmt.Sprintf("Generated %v with id=%v", utils.AsName(utils.PathStrs(mt.Type.Path)...), mt.ID.Int64()))
	tg.F.Type().Id(vGend.Name).Struct(inner...)

	// Generate the supporting encode, decode, and variant functions
	tg.variantGenEncode(v, vGend)
	tg.variantGenDecode(v, vGend)
	tg.variantGenVariant(v, vGend)
	tg.variantGenMarshalJson(v, vGend)

	return vGend, nil
}

// Generate the encode function for a variant.
//
// example output:
//
//	func (ty Result) Encode(encoder scale.Encoder) (err error) {
//		if ty.IsOk {
//			err = encoder.PushByte(0)
//			if err != nil {
//				return err
//			}
//			err = encoder.Encode(ty.AsOkField0)
//			if err != nil {
//				return err
//			}
//			return nil
//		}
//		if ty.IsErr {
//			err = encoder.PushByte(1)
//			if err != nil {
//				return err
//			}
//			err = encoder.Encode(ty.AsErrField0)
//			if err != nil {
//				return err
//			}
//			return nil
//		}
//		return fmt.Errorf("Unrecognized variant")
//	}
func (tg *TypeGenerator) variantGenEncode(v *types.Si1TypeDefVariant, vGend *VariantGend) {
	// IMPORTANT: we only implement encode for the actual type, not the pointer (ty *g.name),
	// because otherwise reflection will fail to see that the object has the method if it's embedded
	// in another type
	// output:
	// func (ty g.name) Encode(encoder scale.Encoder) (err error) {...}
	tg.F.Func().Params(
		jen.Id("ty").Id(vGend.Name),
	).Id("Encode").Params(jen.Id("encoder").Qual(SCALE, "Encoder")).Params(
		jen.Err().Error(),
	).BlockFunc(func(g1 *jen.Group) {
		// for each variant, check if variant
		for i, variant := range v.Variants {
			// This index is not necessarily the index that it appears at in the list
			varI := int(variant.Index)
			g1.If(jen.Id("ty").Dot(vGend.IsVarFields[i].Name)).BlockFunc(func(g2 *jen.Group) {
				// if is variant, encode stuff for variant
				g2.Err().Op("=").Id("encoder").Dot("PushByte").Call(jen.Lit(varI))
				utils.ErrorCheckG(g2)
				for j := range variant.Fields {
					g2.Id("err").Op("=").Id("encoder").Dot("Encode").Call(jen.Id("ty").Dot(vGend.AsVarFields[i][j].Name))
					utils.ErrorCheckG(g2)
				}
				// return ok
				g2.Return(jen.Nil())
			})
		}
		// Didn't hit any variant, return error
		g1.Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("Unrecognized variant")))
	})
}

// Generate the decode function for a variant
//
// example output:
//
//	func (ty *Result) Decode(decoder scale.Decoder) (err error) {
//	 	variant, err := decoder.ReadOneByte()
//	 	if err != nil {
//	 		return err
//	 	}
//	 	switch variant {
//	 	case 0:
//	 		ty.IsOk = true
//	 		err = decoder.Decode(&ty.AsOkField0)
//	 		if err != nil {
//	 			return err
//	 		}
//	 		return
//	 	case 1:
//	 		ty.IsErr = true
//	 		err = decoder.Decode(&ty.AsErrField0)
//	 		if err != nil {
//	 			return err
//	 		}
//	 		return
//	 	default:
//	 		return fmt.Errorf("Unrecognized variant")
//	 	}
//	}
func (tg *TypeGenerator) variantGenDecode(v *types.Si1TypeDefVariant, vGend *VariantGend) {
	// func (ty *g.name) Decode(decoder scale.Decoder) (err error) {...}
	tg.F.Func().Params(
		jen.Id("ty").Op("*").Id(vGend.Name),
	).Id("Decode").Params(jen.Id("decoder").Qual(SCALE, "Decoder")).Params(
		jen.Err().Error(),
	).BlockFunc(func(g1 *jen.Group) {
		// variant, err := decoder.ReadOneByte()
		g1.List(jen.Id("variant"), jen.Err()).Op(":=").Id("decoder").Dot("ReadOneByte").Call()
		utils.ErrorCheckG(g1)
		// switch variant {..}
		// Remember that which variant is encoded in the first byte of the variant
		g1.Switch(jen.Id("variant")).BlockFunc(func(g2 *jen.Group) {
			for i, variant := range v.Variants {
				// This index is not necessarily the index that it appears at in the list
				varI := int(variant.Index)
				g2.Case(jen.Lit(varI)).BlockFunc(func(g3 *jen.Group) {
					// ty.isVariantI = true
					g3.Id("ty").Dot(vGend.IsVarFields[i].Name).Op("=").True()

					isPtr, fieldTy, _ := tg.variantFieldUsesPointer(variant)
					if isPtr {
						g3.Var().Id("tmp").Custom(utils.TypeOpts, fieldTy.Code())
						g3.Err().Op("=").Id("decoder").Dot("Decode").Call(
							jen.Op("&").Id("tmp"),
						)
						utils.ErrorCheckG(g3)
						g3.Id("ty").Dot(vGend.AsVarFields[i][0].Name).Op("=").Op("&").Id("tmp")
					} else {
						// decode remaining fields
						for j := range variant.Fields {
							g3.Err().Op("=").Id("decoder").Dot("Decode").Call(
								jen.Op("&").Id("ty").Dot(vGend.AsVarFields[i][j].Name),
							)
							utils.ErrorCheckG(g3)
						}
					}
					g3.Return()
				})
			}
			g2.Default().Block(jen.Return(jen.Qual("fmt", "Errorf").Call(jen.Lit("Unrecognized variant"))))
		})
	})
}

// Generate the 'Variant' function for a variant. This function takes in the variant, and returns
// the variant index
//
// example output:
//
//	func (ty *Result) Variant() (uint8, error) {
//		if ty.IsOk {
//			return 0, nil
//		}
//		if ty.IsErr {
//			return 1, nil
//		}
//		return 0, fmt.Errorf("No variant detected")
//	}
func (tg *TypeGenerator) variantGenVariant(v *types.Si1TypeDefVariant, vGend *VariantGend) {
	tg.F.Func().Params(
		jen.Id("ty").Op("*").Id(vGend.Name),
	).Id("Variant").Call().Call(jen.Id("uint8"), jen.Error()).BlockFunc(func(g1 *jen.Group) {
		for i, variant := range v.Variants {
			g1.If(jen.Id("ty").Dot(vGend.IsVarFields[i].Name)).Block(
				jen.Return(jen.Lit(int(variant.Index)), jen.Nil()),
			)
		}
		g1.Return(jen.Lit(0), jen.Qual("fmt", "Errorf").Call(jen.Lit("No variant detected")))

	})
}

// Generate the 'MarshalJSON' function for a variant. This function takes in the variant, and returns
// a json version of it, nicely handling variants
//
// example output:
//
//		 func (ty Result) MarshalJSON() ([]byte, error) {
//		 	if ty.IsOk {
//	     m := map[string]interface{}{
//	         "Result::Ok": ty.AsOkField0,
//	         }
//	     }
//		 		return json.Marshal(m)
//		 	}
//		 	if ty.IsErr {
//	     m := map[string]interface{}{
//	         "Result::Err": ty.AsErrField0,
//	     }
//		 		return json.Marshal(m)
//		 	}
//		 	return nil, fmt.Errorf("No variant detected")
//		 }
func (tg *TypeGenerator) variantGenMarshalJson(v *types.Si1TypeDefVariant, vGend *VariantGend) {
	tg.F.Func().Params(
		jen.Id("ty").Id(vGend.Name),
	).Id("MarshalJSON").Call().Call(jen.Index().Byte(), jen.Error()).BlockFunc(func(g1 *jen.Group) {
		for i := range v.Variants {
			fullName := jen.Lit(fmt.Sprintf("%s::%s", vGend.Name, v.Variants[i].Name))
			// if ty.Var
			g1.If(jen.Id("ty").Dot(vGend.IsVarFields[i].Name)).BlockFunc(
				func(g2 *jen.Group) {
					// If there's no associated data, just return the name of the variant
					if len(vGend.AsVarFields[i]) == 0 {
						g2.Return(jen.Qual("encoding/json", "Marshal").Call(fullName))
					} else {
						// Otherwise make a map that has the variant name as a key and associated data as the value
						// m := map[string]interface{}{...}
						g2.Id("m").Op(":=").Map(jen.String()).Interface().Values(jen.DictFunc(
							func(d1 jen.Dict) {
								if len(vGend.AsVarFields[i]) == 1 {
									// If there's only one piece of associated data, set that as the value
									// "Result::Ok": ty.AsOkField0
									d1[fullName] = jen.Id("ty").Dot(vGend.AsVarFields[i][0].Name)
								} else {
									// Otherwise, embed them in a another map[string]interface{}
									d1[fullName] = jen.Map(jen.String()).Interface().Values(jen.DictFunc(
										func(d2 jen.Dict) {
											for j := range vGend.AsVarFields[i] {
												d2[jen.Lit(vGend.AsVarFields[i][j].Name)] = jen.Id("ty").Dot(vGend.AsVarFields[i][j].Name)
											}
										}),
									)
								}
							}),
						)
						g2.Return(jen.Qual("encoding/json", "Marshal").Call(jen.Id("m")))
					}
				},
			)
		}
		g1.Return(jen.Nil(), jen.Qual("fmt", "Errorf").Call(jen.Lit("No variant detected")))
	})
}

// We only make variant fields into pointers if the embedded field is also a variant
// This helps to reduce the size of the most egregious offenders -- events and runtimecalls
func (tg *TypeGenerator) variantFieldUsesPointer(variant types.Si1Variant) (bool, GeneratedType, error) {
	if len(variant.Fields) != 1 {
		return false, nil, nil
	}

	innerType, err := tg.GetType(variant.Fields[0].Type.Int64())
	if err != nil {
		return false, nil, err
	}

	return innerType.MType().Type.Def.IsVariant, innerType, nil
}
