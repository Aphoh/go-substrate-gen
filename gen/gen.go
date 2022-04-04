package gen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/aphoh/go-substrate-gen/tdk"
	"github.com/dave/jennifer/jen"
)

type gend struct {
	id   string
	name string
}

type TypeGenerator struct {
	f         *jen.File
	pkgPath   string
	mtypes    map[string]tdk.MTypeInfo
	generated map[string]gend
}

func NewTypeGenerator(meta *metadata.MetaRoot, pkgPath string) TypeGenerator {
	mtypes := map[string]tdk.MTypeInfo{}
	for _, tdef := range meta.Lookup.Types {
		mtypes[tdef.Id] = tdef.Ty
	}
	f := jen.NewFilePath(pkgPath)
	return TypeGenerator{f: f, pkgPath: pkgPath, mtypes: mtypes, generated: map[string]gend{}}
}

func (tg *TypeGenerator) GetType(id string) (*gend, error) {
	if v, ok := tg.generated[id]; ok {
		return &v, nil
	}
	// gend does not exist, we must generate it

	mt := tg.mtypes[id]
	tn, err := mt.GetTypeName()
	if err != nil {
		return nil, err
	}

	switch tn {
	case tdk.TDKPrimitive:
		prim, err := mt.GetPrimitive()
		if err != nil {
			return nil, err
		}
		return tg.GenPrimitive(prim, id)
	case tdk.TDKArray:
		v, err := mt.GetArray()
		if err != nil {
			return nil, err
		}
		return tg.GenArray(v, id)
	case tdk.TDKComposite:
		v, err := mt.GetComposite()
		if err != nil {
			return nil, err
		}
		return tg.GenComposite(v, id, mt.Path)
	default:
		panic(fmt.Sprintf("Got bad type name=%v", tn))
	}
}
