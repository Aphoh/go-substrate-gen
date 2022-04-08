package callgen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/pal"
	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/dave/jennifer/jen"
)

type CallGenerator struct {
	F      *jen.File
	pallet *pal.Pallet
	tygen  *typegen.TypeGenerator
}

func NewCallGenerator(pkgPath string, pallet *pal.Pallet, tygen *typegen.TypeGenerator) CallGenerator {
	F := jen.NewFilePath(pkgPath)
	return CallGenerator{F: F, pallet: pallet, tygen: tygen}
}

func (cg *CallGenerator) Generate() error {
	gend, err := cg.tygen.GetType(cg.pallet.Calls.TypeId)
	if err != nil {
		return err
	}
	ty := gend.MType().Ty
	if tn, err := ty.GetTypeName(); err != nil {
		return err
	} else if tn != tdk.TDKVariant {
		return fmt.Errorf("Call is not a variant??? %v", ty)
	}

	tdvariant, err := ty.GetVariant()
	if err != nil {
		return err
	}
	for _, variant := range tdvariant.Variants {
		for _, doc := range variant.Docs {
			cg.F.Comment(doc)
		}

		// Get all the arguments to our method
		funcName := utils.AsName("Make", variant.Name, "Call")
		funcArgs := []jen.Code{jen.Id("meta").Op("*").Qual(utils.CTYPES, "Metadata")}
		funcArgNames := []string{}
		var callInd uint32
		for _, field := range variant.Fields {
			fGend, err := cg.tygen.GetType(field.TypeId)
			if err != nil {
				return err
			}
			fieldArgs, fieldArgNames, err := cg.tygen.GenerateArgs(fGend, &callInd)
			funcArgs = append(funcArgs, fieldArgs...)
			funcArgNames = append(funcArgNames, fieldArgNames...)
		}

		cg.F.Func().Id(funcName).Call(funcArgs...).Call(jen.Qual(utils.CTYPES, "Call"), jen.Error()).BlockFunc(func(g *jen.Group) {
			g.ReturnFunc(func(g *jen.Group) {
				// Pass meta and Pallet.func_name first
				innerCallArgs := []jen.Code{jen.Id("meta"), jen.Lit(fmt.Sprintf("%v.%v", cg.pallet.Name, variant.Name))}
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
