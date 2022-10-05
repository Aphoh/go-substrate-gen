package callgen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

type CallGenerator struct {
	F      *jen.File
	pallet *types.PalletMetadataV14
	tygen  *typegen.TypeGenerator
}

func NewCallGenerator(pkgPath string, pallet *types.PalletMetadataV14, tygen *typegen.TypeGenerator) CallGenerator {
	F := jen.NewFilePath(pkgPath)
	return CallGenerator{F: F, pallet: pallet, tygen: tygen}
}

func (cg *CallGenerator) Generate() error {
	baseGend, err := cg.tygen.GetType(cg.pallet.Calls.Type.Int64())
	if err != nil {
		return err
	}
	gend, ok := baseGend.(*typegen.VariantGend)
	if !ok {
		fmt.Printf("Warning: Call type %v for pallet %v is not a variant\n",
			cg.pallet.Calls.Type.Int64(), cg.pallet.Name)
		return nil
	}
	// Runtime call type
	rtc, err := cg.tygen.GetCallType()
	if err != nil {
		return err
	}

	// Index of our pallet's index w/in the generated variant
	runtimeInd, err := rtc.IndOf(uint8(cg.pallet.Index))
	if err != nil {
		return err
	}

	if len(rtc.AsVarFields[runtimeInd]) != 1 {
		return fmt.Errorf("Pallet call (id=%v) has multiple variant fields in runtime call (id=%v)",
			gend.MType().ID, rtc.MType().ID)
	}
	rtcAsVarField := rtc.AsVarFields[runtimeInd][0]
	rtcIsVarField := rtc.IsVarFields[runtimeInd]

	// Already checked it's a variant above
	tdvariant := gend.MType().Type.Def.Variant
	for _, variant := range tdvariant.Variants {
		for _, doc := range variant.Docs {
			cg.F.Comment(string(doc))
		}

		// Get all the arguments to our method
		funcName := utils.AsName("Make", string(variant.Name), "Call")
		funcArgs := []jen.Code{}
		funcArgNames := []string{}
		var callInd uint32
		for _, field := range variant.Fields {
			fGend, err := cg.tygen.GetType(field.Type.Int64())
			if err != nil {
				return err
			}
			fieldArgs, fieldArgNames, err := cg.tygen.GenerateArgs(fGend, &callInd, string(field.Name))
			funcArgs = append(funcArgs, fieldArgs...)
			funcArgNames = append(funcArgNames, fieldArgNames...)
		}

		// Index of this call variant in the pallet call gend
		gendInd, err := gend.IndOf(uint8(variant.Index))
		if err != nil {
			return err
		}

		cg.F.Func().Id(funcName).Call(funcArgs...).Call(rtc.Code()).BlockFunc(func(g1 *jen.Group) {
			g1.ReturnFunc(func(g2 *jen.Group) {
				// return
				g2.Custom(utils.TypeOpts, rtc.Code()).BlockFunc(func(g3 *jen.Group) {
					// RuntimeCall {}
					g3.Id(rtcIsVarField.Name).Op(":").Lit(true).Op(",")
					g3.Id(rtcAsVarField.Name).Op(":").Custom(utils.TypeOpts, gend.Code()).BlockFunc(func(g4 *jen.Group) {
						// PalletCall {}
						g4.Id(gend.IsVarFields[gendInd].Name).Op(":").Lit(true).Op(",")
						for i, fld := range gend.AsVarFields[gendInd] {
							if fld.IsPtr {
								g4.Id(fld.Name).Op(": &").Id(funcArgNames[i]).Op(",")
							} else {
								g4.Id(fld.Name).Op(":").Id(funcArgNames[i]).Op(",")
							}
						}
					}).Op(",")
				})
			})
		})
	}
	return nil
}
