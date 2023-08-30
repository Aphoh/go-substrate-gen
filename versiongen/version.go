package versiongen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

type VersionGenerator struct {
	F       *jen.File
	version types.RuntimeVersion
}

func NewVersionGenerator(pkgPath string, version types.RuntimeVersion) VersionGenerator {
	F := jen.NewFilePath(pkgPath)
	return VersionGenerator{F: F, version: version}
}

func (vg *VersionGenerator) Generate() (string, error) {
	err := vg.generateVersionsCompatible()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%#v", vg.F), nil
}

// Generate a function to check that the version of the connected chain is the same as the chain the
// code was generated for
func (vg *VersionGenerator) generateVersionsCompatible() error {
	args := []jen.Code{jen.Id("state").Qual(utils.GSRPCState, "State")}
	ret := jen.List(jen.Id("ret").Bool(), jen.Err().Error())

	vg.F.Func().Id("VersionsCompatible").Call(args...).Call(ret).BlockFunc(func(g *jen.Group) {
		g.List(jen.Id("vers"), jen.Err()).Op(":=").Id("state").Dot("GetRuntimeVersionLatest").Call()
		utils.ErrorCheckWithNamedArgs(g)
		g.Id("ret").Op("=").Uint32().Call(jen.Id("vers").Dot("SpecVersion")).Op("==").Lit(uint32(vg.version.SpecVersion))
		g.Return()
	})

	return nil
}

// This is what the generated code should look like
// func VersionsCompatible(state state.State) (ret bool, err error) {
// 	vers, err := state.GetRuntimeVersionLatest()
// 	if err != nil {
// 		return
// 	}
// 	ret = vers.SpecVersion == 100
// 	return
// }
