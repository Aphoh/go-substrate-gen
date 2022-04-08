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
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (sg *StorageGenerator) GenPlain(v pal.STPlain, item *pal.SItem, prefix string) error {
	// get inner type
	gend, err := sg.tygen.GetType(string(v))
	if err != nil {
		return err
	}

	args := []jen.Code{jen.Id("meta").Op("*").Qual(utils.CTYPES, "Metadata")} // pointer to metadaa

	parsedType := gend.MType().Ty
	tn, err := parsedType.GetTypeName()
	if err != nil {
		return err
	}

	keyArgNames := []string{}

	isTuple := tn == tdk.TDKTuple
	if !isTuple {
		// Not a tuple, just take the args
		keyArgNames = append(keyArgNames, "args")
		args = append(args, jen.Id("args").Custom(utils.TypeOpts, gend.Code()))
	} else {
		tdef, err := parsedType.GetTuple()
		if err != nil {
			return err
		}
		for i, typeId := range *tdef {
			gend, err := sg.tygen.GetType(typeId)
			if err != nil {
				return err
			}
			argName := fmt.Sprintf("arg%v", i)
			keyArgNames = append(keyArgNames, argName)
			args = append(args, jen.Id(argName).Custom(utils.TypeOpts, gend.Code()))
		}
	}

	//types.CreateStorageKey

	sg.F.Comment(fmt.Sprintf("Make a storage key for %v id=%v", item.Name, v))

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

func (sg *StorageGenerator) GenMap(p pal.STMap, item *pal.SItem, prefix string) error {
	return nil
}
