package typegen

import (
	"fmt"
	"strings"

	"github.com/aphoh/go-substrate-gen/utils"
	types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

type TypeGenerator struct {
	F       *jen.File
	PkgPath string

	// Lazily initialized id for the runtime's call type
	callId *int64

	mtypes      map[int64]types.PortableTypeV14
	generated   map[int64]GeneratedType
	nameCount   map[string]uint32
	namegenOpts map[string]NamegenOpt
}

type NamegenOpt struct {
	fullParams bool
	fullPath   bool
}

func NewTypeGenerator(meta *types.MetadataV14, encodedMetadata string, pkgPath string) TypeGenerator {
	mtypes := map[int64]types.PortableTypeV14{}
	for _, tdef := range meta.Lookup.Types {
		mtypes[tdef.ID.Int64()] = tdef
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

	// Put metadata in the header of the types
	f.Const().Id("encMeta").Op("=").Lit(encodedMetadata)
	f.Var().Id("Meta").Qual(utils.CTYPES, "Metadata")
	f.Var().Id("_").Op("=").Qual(utils.CTYPES, "DecodeFromHexString").Call(jen.Id("encMeta"), jen.Op("&").Id("Meta"))

	return TypeGenerator{F: f, PkgPath: pkgPath, mtypes: mtypes, generated: map[int64]GeneratedType{}, nameCount: map[string]uint32{}, namegenOpts: ng}
}

func (tg *TypeGenerator) GetCallType() (*VariantGend, error) {
	if tg.callId == nil {
		cid, err := getCallTypeId(tg.mtypes)
		if err != nil {
			return nil, err
		}
		tg.callId = &cid
	}

	gend, err := tg.GetType(*tg.callId)
	if err != nil {
		return nil, err
	}
	v, ok := gend.(*VariantGend)
	if !ok {
		return nil, fmt.Errorf("Call (id=%v) is not a variant", *tg.callId)
	}
	return v, nil
}

func (tg *TypeGenerator) MetaCode() *jen.Statement {
	return jen.Qual(tg.PkgPath, "Meta")
}

func (tg *TypeGenerator) GetGenerated() string {
	return fmt.Sprintf("%#v", tg.F)
}

func (tg *TypeGenerator) GetType(id int64) (GeneratedType, error) {
	if v, ok := tg.generated[id]; ok {
		return v, nil
	}
	// gend does not exist, we must generate it

	mt := tg.mtypes[id]
	tdef := mt.Type.Def

	if tdef.IsArray {
		return tg.GenArray(&tdef.Array, &mt)
	} else if tdef.IsBitSequence {
		return tg.GenBitsequence(&tdef.BitSequence, &mt)
	} else if tdef.IsCompact {
		return tg.GenCompact(&tdef.Compact, &mt)
	} else if tdef.IsComposite {
		return tg.GenComposite(&tdef.Composite, &mt)
	} else if tdef.IsSequence {
		return tg.GenSequence(&tdef.Sequence, &mt)
	} else if tdef.IsPrimitive {
		return tg.GenPrimitive(&tdef.Primitive, &mt)
	} else if tdef.IsTuple {
		return tg.GenTuple(&tdef.Tuple, &mt)
	} else if tdef.IsVariant {
		return tg.GenVariant(&tdef.Variant, &mt)
	} else {
		return nil, fmt.Errorf("Got bad type=%v for id=%v\n", tdef, id)
	}
}

// Generates args and they string names from a generated type. This recursively pulls away tuples.
// Index is the starting index for the argument names (e.g. arg1, arg2...)
func (tg *TypeGenerator) GenerateArgs(gend GeneratedType, index *uint32, namePrefixes ...string) ([]jen.Code, []string, error) {
	args := []jen.Code{}
	names := []string{}
	parsedType := gend.MType().Type

	if !parsedType.Def.IsTuple {
		// Not a tuple, just add an argument. Use the index to guarantee uniqueness.
		name := utils.AsArgName(append(namePrefixes, fmt.Sprint(*index))...)

		names = append(names, name)
		args = append(args, jen.Id(name).Custom(utils.TypeOpts, gend.Code()))
		*index += 1
	} else {
		tdef := parsedType.Def.Tuple
		for _, typeId := range tdef {
			gend, err := tg.GetType(typeId.Int64())
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

func getCallTypeId(mtypes map[int64]types.PortableTypeV14) (int64, error) {
	for tyId, ty := range mtypes {
		if len(ty.Type.Path) >= 2 {
			p0 := string(ty.Type.Path[0])
			p1 := string(ty.Type.Path[1])
			// Looking for *_runtime::Call
			if strings.HasSuffix(p0, "_runtime") && p1 == "Call" {
				return tyId, nil
			}
		}
	}
	return 0, fmt.Errorf("No call type found. Expected a path like *_runtime::Call")
}

func reverse(strs []string) (rev []string) {
	for _, s := range strs {
		rev = append(rev, s)
	}
	return
}

func (tg *TypeGenerator) nameFromParams(base []string, params []types.Si1TypeParameter) (string, error) {
	sName := utils.AsName(base...)
	for _, p := range params {
		if p.HasType {
			pgend, err := tg.GetType(p.Type.Int64())
			if err != nil {
				return "", err
			}
			if p.Name != "" {
				base = append(base, string(p.Name))
			}
			base = append(base, pgend.DisplayName())
			sName = utils.AsName(base...)
		}
	}
	return sName, nil
}

func (tg *TypeGenerator) getStructName(mt *types.PortableTypeV14) (string, error) {
	baseName := string(mt.Type.Path[len(mt.Type.Path)-1])
	opts := tg.namegenOpts[baseName]
	// by default only take the last elt of the path (the rust struct name)
	nameWords := []string{baseName}
	if opts.fullPath {
		nameWords = utils.PathStrs(mt.Type.Path)
	}
	sName := utils.AsName(nameWords...)

	var err error
	if opts.fullParams {
		// Get the name with all parameters
		sName, err = tg.nameFromParams(nameWords, mt.Type.Params)
		if err != nil {
			return "", err
		}
	} else {
		// Add params, stopping if its unique
		for i := range mt.Type.Params {
			sName, err = tg.nameFromParams(nameWords, mt.Type.Params[:i])
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
