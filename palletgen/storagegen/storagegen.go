package storagegen

import (
	"encoding/hex"
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/pal"
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

// This is when the stored type is just one thing
func (sg *StorageGenerator) GenPlain(v pal.STPlain, item *pal.SItem, prefix string) error {
	// get inner type
	args := []jen.Code{jen.Id("meta").Op("*").Qual(utils.CTYPES, "Metadata")} // pointer to metadaa

	sg.F.Comment(fmt.Sprintf("Make a storage key for %v id=%v", item.Name, v))
	for _, doc := range item.Docs {
		sg.F.Comment(doc)
	}

	methodName := utils.AsName("Make", item.Name, "StorageKey")
	sg.F.Func().Id(methodName).Call(args...).Call(jen.Qual(utils.CTYPES, "StorageKey"), jen.Error()).BlockFunc(func(g *jen.Group) {
		g.ReturnFunc(func(g1 *jen.Group) {
			g1.Qual(utils.CTYPES, "CreateStorageKey").Call(jen.Id("meta"), jen.Lit(prefix), jen.Lit(item.Name))
		})
	})
	retGend, err := sg.tygen.GetType(string(v))
	if err != nil {
		return err
	}
	sg.generateGetter(true, methodName, args, []string{"meta"}, retGend, item)
	sg.generateGetter(false, methodName, args, []string{"meta"}, retGend, item)
	return nil
}

func (sg *StorageGenerator) GenMap(p pal.STMap, item *pal.SItem, prefix string) error {
	// get inner type
	gend, err := sg.tygen.GetType(p.KeyTypeId)
	if err != nil {
		return err
	}

	args := []jen.Code{jen.Id("meta").Op("*").Qual(utils.CTYPES, "Metadata")} // pointer to metadaa
	var ind uint32 = 0
	newArgs, keyArgNames, err := sg.tygen.GenerateArgs(gend, &ind)
	args = append(args, newArgs...)

	sg.F.Comment(fmt.Sprintf("Make a storage key for %v", item.Name))
	for _, doc := range item.Docs {
		sg.F.Comment(doc)
	}

	methodName := utils.AsName("Make", item.Name, "StorageKey")
	sg.F.Func().Id(methodName).Call(args...).Call(jen.Qual(utils.CTYPES, "StorageKey"), jen.Error()).BlockFunc(func(g *jen.Group) {
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

	retGend, err := sg.tygen.GetType(p.ValueTypeId)
	if err != nil {
		return err
	}
	sKeyArgNames := append([]string{"meta"}, keyArgNames...)
	sg.generateGetter(true, methodName, args, sKeyArgNames, retGend, item)
	sg.generateGetter(false, methodName, args, sKeyArgNames, retGend, item)
	return nil
}

func (sg *StorageGenerator) generateGetter(withBlockhash bool, sKeyMethod string, sKeyArgs []jen.Code, sKeyArgNames []string, returnType typegen.GeneratedType, item *pal.SItem) error {

	args := []jen.Code{jen.Id("state").Op("*").Qual(utils.GSRPCState, "State")}
	if withBlockhash {
		args = append(args, jen.Id("bhash").Qual(utils.CTYPES, "Hash"))
	}
	args = append(args, sKeyArgs...)
	keyArgCode := []jen.Code{}
	for _, name := range sKeyArgNames {
		keyArgCode = append(keyArgCode, jen.Id(name))
	}
	retArgs := []jen.Code{}
	retArgs = append(retArgs, jen.Id("ret").Custom(utils.TypeOpts, returnType.Code()))
	if item.Modifier == "Optional" {
		retArgs = append(retArgs, jen.Id("isSome").Id("bool"))
	}
	ret := jen.List(append(retArgs, jen.Err().Error())...)

	// If default, parse the default result statically
	var defaultBytesName string
	if item.Modifier == "Default" {
		defaultBytesName = utils.AsName(item.Name, "ResultDefaultBytes")
		hexStr := item.Fallback[2:]
		if _, err := hex.DecodeString(hexStr); err != nil {
			return fmt.Errorf("Invalid hex string for item %v, %v", item.Name, item.Fallback)
		}
		if withBlockhash {
			// Don't redefine both for with blockhash and without
			sg.F.Var().List(jen.Id(defaultBytesName), jen.Id("_")).Op("=").Qual("encoding/hex", "DecodeString").Call(jen.Lit(hexStr))
		}
	}

	var methodName string
	if withBlockhash {
		methodName = utils.AsName("Get", item.Name)
	} else {
		methodName = utils.AsName("Get", item.Name, "Latest")
	}

	sg.F.Func().Id(methodName).Call(args...).Call(ret).BlockFunc(func(g *jen.Group) {
		// Get storage key
		g.List(jen.Id("key"), jen.Err()).Op(":=").Id(sKeyMethod).Call(keyArgCode...)
		utils.ErrorCheckWithNamedArgs(g)
		if item.Modifier != "Optional" {
      // if it's optional, this is defined in the return args
			g.Var().Id("isSome").Bool()
		}
		if withBlockhash {
			g.List(jen.Id("isSome"), jen.Err()).Op("=").Id("state").Dot("GetStorage").Call(jen.Id("key"), jen.Op("&").Id("ret"), jen.Id("bhash"))
		} else {
			g.List(jen.Id("isSome"), jen.Err()).Op("=").Id("state").Dot("GetStorageLatest").Call(jen.Id("key"), jen.Op("&").Id("ret"))
		}
		utils.ErrorCheckWithNamedArgs(g)

		if item.Modifier == "Default" {
			// If not optional, return the default when isSome is false
			g.If(jen.Op("!").Id("isSome")).BlockFunc(func(g1 *jen.Group) {
				g1.Err().Op("=").Qual(utils.CTYPES, "DecodeFromBytes").Call(jen.Id(defaultBytesName), jen.Op("&").Id("ret"))
				utils.ErrorCheckWithNamedArgs(g1)
			})
		}
		g.Return()

	})

	return nil
}
