package storagegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/pal"
	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/dave/jennifer/jen"
)

type StorageGenerator struct {
	F         *jen.File
	storage   *pal.Storage
	tygen     *typegen.TypeGenerator
	typesPath string
}

func NewStorageGenerator(pkgName string, storage *pal.Storage, tygen *typegen.TypeGenerator, typesPath string) StorageGenerator {
	F := jen.NewFile(pkgName)
	F.ImportAlias(utils.GSRPC, "gsrpc")
	return StorageGenerator{F: F, storage: storage, tygen: tygen, typesPath: typesPath}
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

	// TODO: should this be just a pattern thingy

	sg.F.Comment(fmt.Sprintf("Make a storage key for %v %v", item.Name, v))

	var argStatement *jen.Statement
	if gend.Global {
		argStatement = jen.Id("args").Op("*").Id(gend.Name)
	} else {
		argStatement = jen.Id("args").Op("*").Qual(sg.typesPath, gend.Name)
	}

	sg.F.Func().Id(utils.AsName("Make", item.Name, "StorageKey")).Call(
		// Function arguments
		jen.Id("meta").Op("*").Qual(utils.CTYPES, "Metadata"), // pointer to metadata
		argStatement,
	).BlockFunc(func(g *jen.Group) {
		// var byteArgs [][]byte
		g.Var().Id("byteArgs").Index().Index().Byte()
		// var err error
		g.Var().Err().Error()
		// Encode the given type.
		g.If( // check if it's a tuple
			jen.List(jen.Id("v"), jen.Id("ok")).Op(":=").Id("args").Op(".").Parens(jen.Qual(sg.typesPath, utils.TupleIface)).Op(";").Id("ok"),
		).BlockFunc(func(g1 *jen.Group) {
			// Set the byte args
			g1.List(jen.Err(), jen.Id("byteArgs")).Op("=").Id("v").Dot(utils.TupleEncodeEach).Call()
			utils.ErrorCheckG(g1)
		}).Else().BlockFunc(func(g1 *jen.Group) {
			g1.List(jen.Id("encBytes"), jen.Err()).Op("=").Qual(utils.CTYPES, "EncodeToBytes").Call(jen.Id("args"))
			utils.ErrorCheckG(g1)
			g1.Id("byteArgs").Op("=").Index().Index().Byte().Values(jen.Id("encBytes"))
		})
		// byteArgs is set, just make the key
		g.ReturnFunc(func(g1 *jen.Group) {
			g1.Qual(utils.CTYPES, "CreateStorageKey").Call(jen.Id("meta"), jen.Lit(prefix), jen.Lit(item.Name), jen.Id("byteArgs").Op("..."))
		})
	})
	return nil
}

func (sg *StorageGenerator) GenMap(p pal.STMap, item *pal.SItem, prefix string) error {
	return nil
}
