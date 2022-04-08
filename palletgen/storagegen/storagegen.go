package storagegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/pal"
	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/dave/jennifer/jen"
)

type StorageGenerator struct {
	F       *jen.File
	storage *pal.Storage
	tygen   *typegen.TypeGenerator
}

func NewStorageGenerator(pkgPath string, storage *pal.Storage, tygen *typegen.TypeGenerator) StorageGenerator {
	F := jen.NewFilePath(pkgPath)
	return StorageGenerator{F: F, storage: storage, tygen: tygen}
}

func (sg *StorageGenerator) Generate() (err error) {
	for _, it := range sg.storage.Items {
		ks := []string{}
		for k := range it.Type {
			ks = append(ks, k)
		}
		if len(ks) != 1 {
			return fmt.Errorf("Incorrect storage type %#v", it.Type)
		}
		switch ks[0] {
		case pal.STKPlain:
			val, err := it.GetTypePlain()
			if err != nil {
				return err
			}
			err = sg.GenPlain(val, &it, sg.storage.Prefix)
		case pal.STKMap:
			val, err := it.GetTypeMap()
			if err != nil {
				return err
			}
			err = sg.GenMap(val, &it, sg.storage.Prefix)
		default:
			return fmt.Errorf("Unsupported storage type %v in %v", ks[0], sg.storage.Prefix)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (sg *StorageGenerator) GenPlain(v pal.STPlain, item *pal.SItem, prefix string) error {
	// get inner type
	args := []jen.Code{jen.Id("meta").Op("*").Qual(utils.CTYPES, "Metadata")} // pointer to metadaa

	sg.F.Comment(fmt.Sprintf("Make a storage key for %v id=%v", item.Name, v))
	for _, doc := range item.Docs {
		sg.F.Comment(doc)
	}

	sg.F.Func().Id(utils.AsName("Make", item.Name, "StorageKey")).Call(args...).Call(jen.Qual(utils.CTYPES, "StorageKey"), jen.Error()).BlockFunc(func(g *jen.Group) {
		g.ReturnFunc(func(g1 *jen.Group) {
			g1.Qual(utils.CTYPES, "CreateStorageKey").Call(jen.Id("meta"), jen.Lit(prefix), jen.Lit(item.Name))
		})
	})
	return nil
}

// Generates args and they string names from a generated type. This recursively pulls away tuples.
// Index is the starting index for the argument names (e.g. arg1, arg2...)
func (sg *StorageGenerator) generateArgs(gend typegen.GeneratedType, index *uint32) ([]jen.Code, []string, error) {
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
			gend, err := sg.tygen.GetType(typeId)
			if err != nil {
				return nil, nil, err
			}
			newArgs, newNames, err := sg.generateArgs(gend, index)
			if err != nil {
				return nil, nil, err
			}

			names = append(names, newNames...)
			args = append(args, newArgs...)
		}
	}
	return args, names, nil
}

func (sg *StorageGenerator) GenMap(p pal.STMap, item *pal.SItem, prefix string) error {
	// get inner type
	gend, err := sg.tygen.GetType(p.KeyTypeId)
	if err != nil {
		return err
	}

	args := []jen.Code{jen.Id("meta").Op("*").Qual(utils.CTYPES, "Metadata")} // pointer to metadaa
	var ind uint32 = 0
	newArgs, keyArgNames, err := sg.generateArgs(gend, &ind)
	args = append(args, newArgs...)

	sg.F.Comment(fmt.Sprintf("Make a storage key for %v", item.Name))
	for _, doc := range item.Docs {
		sg.F.Comment(doc)
	}

	sg.F.Func().Id(utils.AsName("Make", item.Name, "StorageKey")).Call(args...).Call(jen.Qual(utils.CTYPES, "StorageKey"), jen.Error()).BlockFunc(func(g *jen.Group) {
		// byteArgs := [][]byte{}
		g.Id("byteArgs").Op(":=").Index().Index().Byte().Values()
		// encBytes := []byte{}
		g.Id("encBytes").Op(":=").Index().Byte().Values()
		// var err error
		g.Var().Err().Error()
		for _, argName := range keyArgNames {
			g.List(jen.Id("encBytes"), jen.Err()).Op("=").Qual(utils.CTYPES, "EncodeToBytes").Call(jen.Id(argName))
			utils.ErrorCheckWithNil(g)
			g.Id("byteArgs").Op("=").Append(jen.Id("byteArgs"), jen.Id("encBytes"))
		}
		g.ReturnFunc(func(g1 *jen.Group) {
			g1.Qual(utils.CTYPES, "CreateStorageKey").Call(jen.Id("meta"), jen.Lit(prefix), jen.Lit(item.Name), jen.Id("byteArgs").Op("..."))
		})
	})
	return nil
}
