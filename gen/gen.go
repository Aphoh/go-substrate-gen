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
	mtypes    map[string]tdk.MType
	generated map[string]gend
}

func NewTypeGenerator(meta *metadata.MetaRoot, pkgPath string) TypeGenerator {
	mtypes := map[string]tdk.MType{}
	for _, tdef := range meta.Lookup.Types {
		mtypes[tdef.Id] = tdef
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
	tn, err := mt.Ty.GetTypeName()
	if err != nil {
		return nil, err
	}

	switch tn {
	case tdk.TDKArray:
		v, err := mt.Ty.GetArray()
		if err != nil {
			return nil, err
		}
		return tg.GenArray(v, id)
	case tdk.TDKComposite:
		v, err := mt.Ty.GetComposite()
		if err != nil {
			return nil, err
		}
		return tg.GenComposite(v, &mt)
	case tdk.TDKSequence:
		v, err := mt.Ty.GetSequence()
		if err != nil {
			return nil, err
		}
		return tg.GenSequence(v, &mt)
	case tdk.TDKPrimitive:
		prim, err := mt.Ty.GetPrimitive()
		if err != nil {
			return nil, err
		}
		return tg.GenPrimitive(prim, id)
	case tdk.TDKTuple:
		tup, err := mt.Ty.GetTuple()
		if err != nil {
			return nil, err
		}
		return tg.GenTuple(tup, &mt)
	case tdk.TDKVariant:
		v, err := mt.Ty.GetVariant()
		if err != nil {
			return nil, err
		}
		return tg.GenVariant(v, &mt)
	default:
		return nil, fmt.Errorf("Got bad type name=%v\n", tn)
	}
}

func (tg *TypeGenerator) GenAll() (string, error) {
	for id := range tg.mtypes {
		if _, err := tg.GetType(id); err != nil {
			println("Got error getting type", "type", id, "err", err.Error())
		}
	}
	return fmt.Sprintf("%#v", tg.f), nil
}
