package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	types "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

// The type generator is used to generate all types necessary to support the storage accesses and
// extrinsics provided by the pallets of the substrate chain. It gradually builds a mapping from IDs
// (indexes within the metadata types list) to generated types. It is also used to generate unique
// names for certain purposes.
type TypeGenerator struct {
	// A jen file which every generated type gets written to
	F *jen.File
	// The path to the 'types' path in which to generate all types
	PkgPath string

	// Lazily initialized id for the runtime's call type
	// This is used to convert extrinsics into actual runnable calls in the client
	callId *int64

	// A map from ID -> go-rpc-types
	mtypes map[int64]types.PortableTypeV14
	// A map from ID -> our generated type code
	generated map[int64]GeneratedType
	// A map from name -> the number of uses. This is used to ensure unique struct names by
	// appending numbers to the end of names.
	nameCount map[string]uint32
	// A map used to keep track of which rust types should be named based on their full path or full
	// parameters, instead of stopping at the first unique part
	namegenOpts map[string]NamegenOpt
}

// Options for name generation (ideally for a particular group of rust types)
type NamegenOpt struct {
	// Generate the name based on the full set of generics used in that rust type, instead of
	// stopping at the first unique name
	fullParams bool
	// Generate the name based on the full path of that rust type, instead of only using the last
	// element of the path
	fullPath bool
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

	// Put metadata in the header of the types, used for creating storage keys correctly
	f.Const().Id("encMeta").Op("=").Lit(encodedMetadata)
	f.Var().Id("Meta").Qual(utils.CTYPES, "Metadata")
	f.Var().Id("_").Op("=").Qual(utils.CCODEC, "DecodeFromHex").Call(jen.Id("encMeta"), jen.Op("&").Id("Meta"))

	return TypeGenerator{F: f, PkgPath: pkgPath, mtypes: mtypes, generated: map[int64]GeneratedType{}, nameCount: map[string]uint32{}, namegenOpts: ng}
}

// Get a jen statement for the metadata of the chain. This is used to create the correct storage key
// when calling into go-substrate-rpc-client
func (tg *TypeGenerator) MetaCode() *jen.Statement {
	return jen.Qual(tg.PkgPath, "Meta")
}

// Return a string representation of all
func (tg *TypeGenerator) GetGenerated() string {
	return fmt.Sprintf("%#v", tg.F)
}

// Returns the generated type for the given id, generating it if it did not previously exist
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
		return nil, fmt.Errorf("got bad type=%v for id=%v", tdef, id)
	}
}

// Generates args and their string names from a generated type. This recursively pulls away tuples.
// Index is the starting index for the argument names (e.g. arg1, arg2...)
// Tuples like (typeA, (typeB, ())) will be flattened into (typeA, typeB)
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

// Generate all types. This should not be used outside of testing
func (tg *TypeGenerator) GenAll() (string, error) {
	for id := range tg.mtypes {
		if _, err := tg.GetType(id); err != nil {
			println("Got error getting type", "type", id, "err", err.Error())
		}
	}
	return fmt.Sprintf("%#v", tg.F), nil
}

// Generate a name based on the params (rust generics) of the type. These names may not be unique
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

// Generate and return a unique name for a struct, based on its path and parameters, as well as any
// applicable nameOpt
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
		// Add params (generics in rust), stopping if its unique
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
