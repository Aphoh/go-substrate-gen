package storagegen

import (
	"encoding/hex"
	"fmt"

	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

type StorageGenerator struct {
	F       *jen.File
	storage *types.StorageMetadataV14
	tygen   *typegen.TypeGenerator
}

func NewStorageGenerator(pkgPath string, storage *types.StorageMetadataV14, tygen *typegen.TypeGenerator) StorageGenerator {
	F := jen.NewFilePath(pkgPath)
	return StorageGenerator{F: F, storage: storage, tygen: tygen}
}

func (sg *StorageGenerator) Generate() (err error) {
	for _, it := range sg.storage.Items {
		//ks := []string{}
		//for k := range it.Type.AsMap {
		//	ks = append(ks, k)
		//}
		//if len(ks) != 1 {
		//	return fmt.Errorf("Incorrect storage type %#v", it.Type)
		//}
		if it.Type.IsPlainType {
			err = sg.GenPlain(it.Type.AsPlainType, &it, string(sg.storage.Prefix))
		} else if it.Type.IsMap {
			err = sg.GenMap(it.Type.AsMap, &it, string(sg.storage.Prefix))
		} else {
			return fmt.Errorf("Unsupported storage type %v in %v", it, sg.storage.Prefix)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// This is when the stored type is just one thing
func (sg *StorageGenerator) GenPlain(v types.Si1LookupTypeID, item *types.StorageEntryMetadataV14, prefix string) error {
	// get inner type
	args := []jen.Code{} // pointer to metadaa

	sg.F.Comment(fmt.Sprintf("Make a storage key for %v id=%v", item.Name, v))
	for _, doc := range item.Documentation {
		sg.F.Comment(string(doc))
	}

	methodName := utils.AsName("Make", string(item.Name), "StorageKey")
	sg.F.Func().Id(methodName).Call(args...).Call(jen.Qual(utils.CTYPES, "StorageKey"), jen.Error()).BlockFunc(func(g *jen.Group) {
		g.ReturnFunc(func(g1 *jen.Group) {
			metaArg := jen.Op("&").Custom(utils.TypeOpts, sg.tygen.MetaCode())
			g1.Qual(utils.CTYPES, "CreateStorageKey").Call(metaArg, jen.Lit(prefix), jen.Lit(string(item.Name)))
		})
	})
	retGend, err := sg.tygen.GetType(v.Int64())
	if err != nil {
		return err
	}
	sg.generateGetter(true, methodName, args, []string{}, retGend, item)
	sg.generateGetter(false, methodName, args, []string{}, retGend, item)
	return nil
}

func (sg *StorageGenerator) GenMap(p types.MapTypeV14, item *types.StorageEntryMetadataV14, prefix string) error {
	// get inner type
	gend, err := sg.tygen.GetType(p.Key.Int64())
	if err != nil {
		return err
	}

	args := []jen.Code{} // pointer to metadaa
	var ind uint32 = 0
	// Don't have a field name, so we prefix by the type name
	newArgs, keyArgNames, err := sg.tygen.GenerateArgs(gend, &ind, gend.DisplayName())
	args = append(args, newArgs...)

	sg.F.Comment(fmt.Sprintf("Make a storage key for %v", item.Name))
	for _, doc := range item.Documentation {
		sg.F.Comment(string(doc))
	}

	methodName := utils.AsName("Make", string(item.Name), "StorageKey")
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
			metaArg := jen.Op("&").Custom(utils.TypeOpts, sg.tygen.MetaCode())
			g1.Qual(utils.CTYPES, "CreateStorageKey").Call(metaArg, jen.Lit(prefix), jen.Lit(string(item.Name)), jen.Id("byteArgs").Op("..."))
		})
	})

	retGend, err := sg.tygen.GetType(p.Value.Int64())
	if err != nil {
		return err
	}

	sg.generateGetter(true, methodName, args, keyArgNames, retGend, item)
	sg.generateGetter(false, methodName, args, keyArgNames, retGend, item)
	return nil
}

func (sg *StorageGenerator) generateGetter(withBlockhash bool, sKeyMethod string, sKeyArgs []jen.Code, sKeyArgNames []string, returnType typegen.GeneratedType, item *types.StorageEntryMetadataV14) error {

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
	if item.Modifier.IsOptional {
		retArgs = append(retArgs, jen.Id("isSome").Id("bool"))
	}
	ret := jen.List(append(retArgs, jen.Err().Error())...)

	// If default, parse the default result statically
	var defaultBytesName string
	if item.Modifier.IsDefault {
		defaultBytesName = utils.AsName(string(item.Name), "ResultDefaultBytes")
		hexStr := hex.EncodeToString(item.Fallback)
		if withBlockhash {
			// Don't redefine both for with blockhash and without
			sg.F.Var().List(jen.Id(defaultBytesName), jen.Id("_")).Op("=").Qual("encoding/hex", "DecodeString").Call(jen.Lit(hexStr))
		}
	}

	var methodName string
	if withBlockhash {
		methodName = utils.AsName("Get", string(item.Name))
	} else {
		methodName = utils.AsName("Get", string(item.Name), "Latest")
	}

	sg.F.Func().Id(methodName).Call(args...).Call(ret).BlockFunc(func(g *jen.Group) {
		// Get storage key
		g.List(jen.Id("key"), jen.Err()).Op(":=").Id(sKeyMethod).Call(keyArgCode...)
		utils.ErrorCheckWithNamedArgs(g)
		if !item.Modifier.IsOptional {
			// if it's optional, this is defined in the return args
			g.Var().Id("isSome").Bool()
		}
		if withBlockhash {
			g.List(jen.Id("isSome"), jen.Err()).Op("=").Id("state").Dot("GetStorage").Call(jen.Id("key"), jen.Op("&").Id("ret"), jen.Id("bhash"))
		} else {
			g.List(jen.Id("isSome"), jen.Err()).Op("=").Id("state").Dot("GetStorageLatest").Call(jen.Id("key"), jen.Op("&").Id("ret"))
		}
		utils.ErrorCheckWithNamedArgs(g)

		if item.Modifier.IsDefault {
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
