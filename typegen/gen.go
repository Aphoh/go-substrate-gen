package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/dave/jennifer/jen"
)

type TypeGenerator struct {
	F           *jen.File
	PkgPath     string
	mtypes      map[string]tdk.MType
	generated   map[string]GeneratedType
	nameCount   map[string]uint32
	namegenOpts map[string]NamegenOpt
}

type NamegenOpt struct {
	fullParams bool
	fullPath   bool
}

func NewTypeGenerator(meta *metadata.MetaRoot, pkgPath string) TypeGenerator {
	mtypes := map[string]tdk.MType{}
	for _, tdef := range meta.Lookup.Types {
		mtypes[tdef.Id] = tdef
	}
	f := jen.NewFilePath(pkgPath)
	// Public, Event, Error, Call, Signature <- full path
	// Option, WeakBoundedVec, BoundedVec, BTreeMap <- Full params
	// TODO: take in config?
	ng := map[string]NamegenOpt{
		"Public":         {fullPath: true},
		"Event":          {fullPath: true},
		"Error":          {fullPath: true},
		"Call":           {fullPath: true},
		"Signature":      {fullPath: true},
		"Option":         {fullParams: true},
		"WeakBoundedVec": {fullParams: true},
		"BoundedVec":     {fullParams: true},
		"BTreeMap":       {fullParams: true},
	}
	return TypeGenerator{F: f, PkgPath: pkgPath, mtypes: mtypes, generated: map[string]GeneratedType{}, nameCount: map[string]uint32{}, namegenOpts: ng}
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

func reverse(strs []string) (rev []string) {
	for _, s := range strs {
		rev = append(rev, s)
	}
	return
}

func (tg *TypeGenerator) nameFromParams(base []string, params []tdk.MTypeParam) (string, error) {
	sName := utils.AsName(base...)
	for _, p := range params {
		if p.Type != nil {
			pgend, err := tg.GetType(*p.Type)
			if err != nil {
				return "", err
			}
			if p.Name != "" {
				base = append(base, p.Name)
			}
			base = append(base, pgend.DisplayName())
			sName = utils.AsName(base...)
		}
	}
	return sName, nil
}

func (tg *TypeGenerator) getStructName(mt *tdk.MType) (string, error) {
	baseName := mt.Ty.Path[len(mt.Ty.Path)-1]
	opts := tg.namegenOpts[baseName]
	// by default only take the last elt of the path (the rust struct name)
	nameWords := []string{baseName}
	if opts.fullPath {
		nameWords = mt.Ty.Path
	}
	sName := utils.AsName(nameWords...)

	var err error
	if opts.fullParams {
		// Get the name with all parameters
		sName, err = tg.nameFromParams(nameWords, mt.Ty.Params)
		if err != nil {
			return "", err
		}
	} else {
		// Add params, stopping if its unique
		for i := range mt.Ty.Params {
			sName, err = tg.nameFromParams(nameWords, mt.Ty.Params[:i])
			if err != nil {
				return "", err
			}
			if tg.nameCount[sName] == 0 {
				break
			}
		}
	}

	// Even with params/path, this name scheme is not unique, so we may have to add an integer postfix
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
func (tg *TypeGenerator) GenerateArgs(gend GeneratedType, index *uint32, namePrefixes ...string) ([]jen.Code, []string, error) {
	args := []jen.Code{}
	names := []string{}
	parsedType := gend.MType().Ty
	tn, err := parsedType.GetTypeName()
	if err != nil {
		return nil, nil, err
	}

	if tn != tdk.TDKTuple {
		// Not a tuple, just add an argument. Use the index to guarantee uniqueness.
		name := utils.AsArgName(append(namePrefixes, fmt.Sprint(*index))...)

		names = append(names, name)
		if gend.IsPrimitive() {
			args = append(args, jen.Id(name).Custom(utils.TypeOpts, gend.Code()))
		} else {
			// Use a pointer if it's not primitive
			args = append(args, jen.Id(name).Op("*").Custom(utils.TypeOpts, gend.Code()))
		}
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
			newArgs, newNames, err := tg.GenerateArgs(gend, index, namePrefixes...)
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
