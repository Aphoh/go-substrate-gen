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
	gend, err := cg.tygen.GetType(cg.pallet.Calls.Type.Int64())
	if err != nil {
		return err
	}
	ty := gend.MType().Type
	if !ty.Def.IsVariant {
		return fmt.Errorf("Call is not a variant??? %v", ty)
	}

	tdvariant := ty.Def.Variant
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

		cg.F.Func().Id(funcName).Call(funcArgs...).Call(jen.Qual(utils.CTYPES, "Call"), jen.Error()).BlockFunc(func(g *jen.Group) {
			g.ReturnFunc(func(g *jen.Group) {
				// Pass meta and Pallet.func_name first
				metaArg := jen.Op("&").Custom(utils.TypeOpts, cg.tygen.MetaCode())
				innerCallArgs := []jen.Code{metaArg, jen.Lit(fmt.Sprintf("%v.%v", cg.pallet.Name, variant.Name))}
				for _, argName := range funcArgNames {
					innerCallArgs = append(innerCallArgs, jen.Id(argName))
				}
				// Call with all passed args
				g.Qual(utils.CTYPES, "NewCall").Call(innerCallArgs...)
			})
		})
	}
	return nil
}
