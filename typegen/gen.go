package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/dave/jennifer/jen"
)

const CTypes = "github.com/centrifuge/go-substrate-rpc-client/v4/types"

type gend struct {
	id   string
	name string
}

type TypeGenerator struct {
	f         *jen.File
	pkgPath   string
	mtypes    map[string]tdk.MType
	generated map[string]gend
	nameCount map[string]uint32
}

func NewTypeGenerator(meta *metadata.MetaRoot, pkgPath string) TypeGenerator {
	mtypes := map[string]tdk.MType{}
	for _, tdef := range meta.Lookup.Types {
		mtypes[tdef.Id] = tdef
	}
	f := jen.NewFilePath(pkgPath)
	f.ImportAlias("github.com/centrifuge/go-substrate-rpc-client/types", "ctypes")
	f.Comment("Add dummy variable so jennifer keeps the imports")
	f.Var().Id("_").Op("=").Qual(CTypes, "NewU8").Call(jen.Lit(0))
	return TypeGenerator{f: f, pkgPath: pkgPath, mtypes: mtypes, generated: map[string]gend{}, nameCount: map[string]uint32{}}
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
	case tdk.TDKCompact:
		v, err := mt.Ty.GetCompact()
		if err != nil {
			return nil, err
		}
		return tg.GenCompact(v, &mt)
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

func (tg *TypeGenerator) getStructName(mt *tdk.MType, base ...string) (*gend, error) {
	nameParams := append(mt.Ty.Path, base...)
	sName := asName(nameParams...)
  // Add params, stopping if its unique
	if tg.nameCount[sName] != 0 {
		for _, p := range mt.Ty.Params {
			if p.Type != nil {
				pgend, err := tg.GetType(*p.Type)
				if err != nil {
					return nil, err
				}
				if p.Name != "" {
					nameParams = append(nameParams, p.Name)
				}
				nameParams = append(nameParams, pgend.name)
        sName = asName(nameParams...)
        if tg.nameCount[sName] == 0 {
          break
        }
			}
		}
	}

	// This name scheme is not unique, so we may have to add an integer postfix
	sName = asName(nameParams...)
	if tg.nameCount[sName] == 0 {
		tg.nameCount[sName] = 1
	} else {
		tg.nameCount[sName] += 1
		sName = asName(sName, fmt.Sprint(tg.nameCount[sName] - 1))
	}
	g := gend{
		id:   mt.Id,
		name: sName,
	}
	tg.generated[mt.Id] = g

	return &g, nil
}

func (tg *TypeGenerator) GenAll() (string, error) {
	for id := range tg.mtypes {
		if _, err := tg.GetType(id); err != nil {
			println("Got error getting type", "type", id, "err", err.Error())
		}
	}
	return fmt.Sprintf("%#v", tg.f), nil
}
