package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/dave/jennifer/jen"
)

type TypeGenerator struct {
	F         *jen.File
	PkgPath   string
	mtypes    map[string]tdk.MType
	generated map[string]GeneratedType
	nameCount map[string]uint32
}

func NewTypeGenerator(meta *metadata.MetaRoot, pkgPath string) TypeGenerator {
	mtypes := map[string]tdk.MType{}
	for _, tdef := range meta.Lookup.Types {
		mtypes[tdef.Id] = tdef
	}
	f := jen.NewFilePath(pkgPath)
	f.Type().Id(utils.TupleIface).Interface(jen.Id(utils.TupleEncodeEach).Call().Index().Index().Byte())
	return TypeGenerator{F: f, PkgPath: pkgPath, mtypes: mtypes, generated: map[string]GeneratedType{}, nameCount: map[string]uint32{}}
}

func (tg *TypeGenerator) GetGenerated() string {
  return fmt.Sprintf("%#v", tg.F)
}

func (tg *TypeGenerator) GetType(id string) (GeneratedType, error) {
	if v, ok := tg.generated[id]; ok {
		return v, nil
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
		return tg.GenArray(v, &mt)
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
		return tg.GenPrimitive(prim, &mt)
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

func (tg *TypeGenerator) getStructName(mt *tdk.MType, base ...string) (string, error) {
	nameParams := append(mt.Ty.Path, base...)
	sName := utils.AsName(nameParams...)
	// Add params, stopping if its unique
	if tg.nameCount[sName] != 0 {
		for _, p := range mt.Ty.Params {
			if p.Type != nil {
				pgend, err := tg.GetType(*p.Type)
				if err != nil {
					return "", err
				}
				if p.Name != "" {
					nameParams = append(nameParams, p.Name)
				}
				nameParams = append(nameParams, pgend.DisplayName())
				sName = utils.AsName(nameParams...)
				if tg.nameCount[sName] == 0 {
					break
				}
			}
		}
	}

	// Even with params, this name scheme is not unique, so we may have to add an integer postfix
	sName = utils.AsName(nameParams...)
	if tg.nameCount[sName] == 0 {
		tg.nameCount[sName] = 1
	} else {
		tg.nameCount[sName] += 1
		sName = utils.AsName(sName, fmt.Sprint(tg.nameCount[sName]-1))
	}
	return sName, nil
}

// Generates args and they string names from a generated type. This recursively pulls away tuples.
// Index is the starting index for the argument names (e.g. arg1, arg2...)
func (tg *TypeGenerator) GenerateArgs(gend GeneratedType, index *uint32) ([]jen.Code, []string, error) {
	args := []jen.Code{}
	names := []string{}
	parsedType := gend.MType().Ty
	tn, err := parsedType.GetTypeName()
	if err != nil {
		return nil, nil, err
	}

	if tn != tdk.TDKTuple {
		// Not a tuple, just take the args
		name := fmt.Sprintf("arg%v", *index)

		names = append(names, name)
		args = append(args, jen.Id(name).Custom(utils.TypeOpts, gend.Code()))
		*index += 1
	} else {
		tdef, err := parsedType.GetTuple()
		if err != nil {
			return nil, nil, err
		}
		for _, typeId := range *tdef {
			gend, err := tg.GetType(typeId)
			if err != nil {
				return nil, nil, err
			}
			newArgs, newNames, err := tg.GenerateArgs(gend, index)
			if err != nil {
				return nil, nil, err
			}

			names = append(names, newNames...)
			args = append(args, newArgs...)
		}
	}
	return args, names, nil
}

func (tg *TypeGenerator) GenAll() (string, error) {
	for id := range tg.mtypes {
		if _, err := tg.GetType(id); err != nil {
			println("Got error getting type", "type", id, "err", err.Error())
		}
	}
	return fmt.Sprintf("%#v", tg.F), nil
}
