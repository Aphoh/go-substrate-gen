package storagegen

import (
	"encoding/hex"
	"fmt"

	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

// The StorageGenerator generates all methods necessary to access a pallet's storage.
// For each item in the storage, it generates:
// - a method to make a storage key (this is public, but not necessary to use)
// - a method to access the storage at a specific block hash
// - a method to access the current storage state
type StorageGenerator struct {
	F       *jen.File
	storage *types.StorageMetadataV14
	tygen   *typegen.TypeGenerator
}

func NewStorageGenerator(pkgPath string, storage *types.StorageMetadataV14, tygen *typegen.TypeGenerator) StorageGenerator {
	F := jen.NewFilePath(pkgPath)
	return StorageGenerator{F: F, storage: storage, tygen: tygen}
}

// Generate storage access methods for all items in storage
func (sg *StorageGenerator) Generate() (err error) {
	for _, it := range sg.storage.Items {
		if it.Type.IsPlainType {
			err = sg.GenPlain(it.Type.AsPlainType, &it, string(sg.storage.Prefix))
		} else if it.Type.IsMap {
			err = sg.GenMap(it.Type.AsMap, &it, string(sg.storage.Prefix))
		} else {
			return fmt.Errorf("unsupported storage type %v in %v", it, sg.storage.Prefix)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// This is when the stored type is just one thing
// This corresponds to the 'StorageValue' in substrate
func (sg *StorageGenerator) GenPlain(v types.Si1LookupTypeID, item *types.StorageEntryMetadataV14, prefix string) error {
	// Add all documentation to the top of the Make...StorageKey method
	sg.F.Comment(fmt.Sprintf("Make a storage key for %v id=%v", item.Name, v))
	for _, doc := range item.Documentation {
		sg.F.Comment(string(doc))
	}

	// example output:
	// func MakeStateStorageKey() (types.StorageKey, error) {
	//   return types.CreateStorageKey(&types1.Meta, "Grandpa", "State")
	// }
	methodName := utils.AsName("Make", string(item.Name), "StorageKey")
	sg.F.Func().Id(methodName).Call().Call(jen.Qual(utils.CTYPES, "StorageKey"), jen.Error()).BlockFunc(func(g *jen.Group) {
		g.ReturnFunc(func(g1 *jen.Group) {
			metaArg := jen.Op("&").Custom(utils.TypeOpts, sg.tygen.MetaCode())
			g1.Qual(utils.CTYPES, "CreateStorageKey").Call(metaArg, jen.Lit(prefix), jen.Lit(string(item.Name)))
		})
	})
	retGend, err := sg.tygen.GetType(v.Int64())
	if err != nil {
		return err
	}

	// Note that the getters need no real arguments, because this is not a map
	sg.generateGetter(true, methodName, []jen.Code{}, []string{}, retGend, item)
	sg.generateGetter(false, methodName, []jen.Code{}, []string{}, retGend, item)
	return nil
}

// Generate the code for a storage item that is a map
// This corresponds to both 'StorageMap', 'StorageDoubleMap' and 'StorageNMap' in substrate
func (sg *StorageGenerator) GenMap(p types.MapTypeV14, item *types.StorageEntryMetadataV14, prefix string) error {
	// get inner type
	gend, err := sg.tygen.GetType(p.Key.Int64())
	if err != nil {
		return err
	}

	// Generate the arguments needed to specify a storage value within a map
	args := []jen.Code{} // pointer to metadata
	// Don't have a field name, so we prefix by the type name
	var ind uint32 = 0
	newArgs, keyArgNames, err := sg.tygen.GenerateArgs(gend, &ind, gend.DisplayName())
	if err != nil {
		return err
	}
	args = append(args, newArgs...)

	sg.F.Comment(fmt.Sprintf("Make a storage key for %v", item.Name))
	for _, doc := range item.Documentation {
		sg.F.Comment(string(doc))
	}

	// The 'make storage key' method creates an array of arguments (byteargs) by encoding each
	// argument with codec before adding it to the array Finally, it returns a storage key created
	// from the go-substrate-rpc-client with the byte arguments and correct types.
	// example output:
	// func MakeSpacesStorageKey(byteArray0 [10]byte) (types.StorageKey, error) {
	//   byteArgs := [][]byte{}
	//   encBytes := []byte{}
	//   var err error
	//   encBytes, err = codec.Encode(byteArray0)
	//   if err != nil {
	//     return nil, err
	//   }
	//   byteArgs = append(byteArgs, encBytes)
	//   return types.CreateStorageKey(&types1.Meta, "Spaces", "Spaces", byteArgs...)
	// }
	methodName := utils.AsName("Make", string(item.Name), "StorageKey")
	sg.F.Func().Id(methodName).Call(args...).Call(jen.Qual(utils.CTYPES, "StorageKey"), jen.Error()).BlockFunc(func(g *jen.Group) {
		// byteArgs := [][]byte{}
		g.Id("byteArgs").Op(":=").Index().Index().Byte().Values()
		// encBytes := []byte{}
		g.Id("encBytes").Op(":=").Index().Byte().Values()
		// var err error
		g.Var().Err().Error()
		for _, argName := range keyArgNames {
			g.List(jen.Id("encBytes"), jen.Err()).Op("=").Qual(utils.CCODEC, "Encode").Call(jen.Id(argName))
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

	// Note that the getter functions here *do* need the arguments provided because the storage item is a map
	sg.generateGetter(true, methodName, args, keyArgNames, retGend, item)
	sg.generateGetter(false, methodName, args, keyArgNames, retGend, item)
	return nil
}

// Generate a getter function for a storage item. If `withBlockHash`, add an argument to get it at a particular block hash and name the function latest
func (sg *StorageGenerator) generateGetter(withBlockhash bool, sKeyMethod string, sKeyArgs []jen.Code, sKeyArgNames []string, returnType typegen.GeneratedType, item *types.StorageEntryMetadataV14) error {

	// Add state and (maybe) blockhash to the method arguments
	args := []jen.Code{jen.Id("state").Qual(utils.GSRPCState, "State")}
	if withBlockhash {
		args = append(args, jen.Id("bhash").Qual(utils.CTYPES, "Hash"))
	}
	args = append(args, sKeyArgs...)

	// Get arguments for the call to Make{..}StorageKey
	keyArgCode := []jen.Code{}
	for _, name := range sKeyArgNames {
		keyArgCode = append(keyArgCode, jen.Id(name))
	}

	// Generate the statement for the return args in the function header
	retArgs := []jen.Code{}
	retArgs = append(retArgs, jen.Id("ret").Custom(utils.TypeOpts, returnType.Code()))
	if item.Modifier.IsOptional {
		retArgs = append(retArgs, jen.Id("isSome").Id("bool"))
	}
	ret := jen.List(append(retArgs, jen.Err().Error())...)

	// If there is a default, parse the default result statically
	// This will output a constant before the getter that then gets referenced in the getter
	// example output:
	// var ContentObjectMetasResultDefaultBytes, _ = hex.DecodeString("0000000000")
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
		// Get storage key from the Make{..}StorageKey method we defined earlier
		g.List(jen.Id("key"), jen.Err()).Op(":=").Id(sKeyMethod).Call(keyArgCode...)
		utils.ErrorCheckWithNamedArgs(g)
		if !item.Modifier.IsOptional {
			// if it's optional, this is defined in the return args
			g.Var().Id("isSome").Bool()
		}

		// Make the actual storage call
		if withBlockhash {
			g.List(jen.Id("isSome"), jen.Err()).Op("=").Id("state").Dot("GetStorage").Call(jen.Id("key"), jen.Op("&").Id("ret"), jen.Id("bhash"))
		} else {
			g.List(jen.Id("isSome"), jen.Err()).Op("=").Id("state").Dot("GetStorageLatest").Call(jen.Id("key"), jen.Op("&").Id("ret"))
		}
		utils.ErrorCheckWithNamedArgs(g)

		// If not optional, return the default when isSome is false
		if item.Modifier.IsDefault {
			g.If(jen.Op("!").Id("isSome")).BlockFunc(func(g1 *jen.Group) {
				g1.Err().Op("=").Qual(utils.CCODEC, "Decode").Call(jen.Id(defaultBytesName), jen.Op("&").Id("ret"))
				utils.ErrorCheckWithNamedArgs(g1)
			})
		}
		g.Return()

	})

	return nil
}
