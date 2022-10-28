package callgen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

// The call generator generates one method per available extrinsic in the pallet which calls the
// corresponding extrinsic.
type CallGenerator struct {
	F      *jen.File
	pallet *types.PalletMetadataV14
	tygen  *typegen.TypeGenerator
}

func NewCallGenerator(pkgPath string, pallet *types.PalletMetadataV14, tygen *typegen.TypeGenerator) CallGenerator {
	F := jen.NewFilePath(pkgPath)
	return CallGenerator{F: F, pallet: pallet, tygen: tygen}
}

// Generate all extrinsic calls for a particular pallet.
// Each is of the form Make{PalletExtrinsicName}Call
func (cg *CallGenerator) Generate() error {
	// Get the base call type for the calls in this pallet
	// Example name:
	// FrameSystemPalletCall
	baseGend, err := cg.tygen.GetType(cg.pallet.Calls.Type.Int64())
	if err != nil {
		return err
	}
	// Cast the base call type as a variant, which it should always be
	gend, ok := baseGend.(*typegen.VariantGend)
	if !ok {
		fmt.Printf("Warning: Call type %v for pallet %v is not a variant\n",
			cg.pallet.Calls.Type.Int64(), cg.pallet.Name)
		return nil
	}

	// Get the runtime call type, which is a variant containing all pallet's calls
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
	// Find the correct fields we need to fill into our runtime call struct
	rtcAsVarField := rtc.AsVarFields[runtimeInd][0]
	rtcIsVarField := rtc.IsVarFields[runtimeInd]

	// Already checked it's a variant above
	tdvariant := gend.MType().Type.Def.Variant

	// Each variant of our pallet call type corresponds to a particular extrinsic of our pallet, so
	// we generate a call for each
	for _, variant := range tdvariant.Variants {
		cg.generateCall(variant, gend, rtc, rtcIsVarField.Name, rtcAsVarField.Name)
	}
	return nil
}

// Generate a function to call a particular pallet extrinsic.
// example output (docs omitted):
// func MakeSetKeyCall(new0 types.MultiAddress) types.TemplateRuntimeCall {
//   return types.TemplateRuntimeCall{
// 	   IsSudo: true,
// 	   AsSudoField0: types.PalletSudoPalletCall{
// 		 IsSetKey:     true,
// 		 AsSetKeyNew0: new0,
// 	   },
// 	 }
// }
func (cg *CallGenerator) generateCall(variant types.Si1Variant, gend, rtc *typegen.VariantGend, rtcIsVarName, rtcAsVarName string) error {
	for _, doc := range variant.Docs {
		cg.F.Comment(string(doc))
	}

	// Get all the arguments to our method
	funcName := utils.AsName("Make", string(variant.Name), "Call")
	funcArgs := []jen.Code{}
	funcArgNames := []string{}
	var callInd uint32
	// The fields on the variants are the arguments to the extrinsic, so we get each
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

	// Generate the actual code for the function
	cg.F.Func().Id(funcName).Call(funcArgs...).Call(rtc.Code()).BlockFunc(func(g1 *jen.Group) {
		g1.ReturnFunc(func(g2 *jen.Group) {
			// return
			g2.Custom(utils.TypeOpts, rtc.Code()).BlockFunc(func(g3 *jen.Group) {
				// RuntimeCall {}
				g3.Id(rtcIsVarName).Op(":").Lit(true).Op(",")
				g3.Id(rtcAsVarName).Op(":").Custom(utils.TypeOpts, gend.Code()).BlockFunc(func(g4 *jen.Group) {
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

	return nil
}
