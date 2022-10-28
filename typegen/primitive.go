package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

//	IsBool = 0
//	IsChar = 1
//	IsStr  = 2
//	IsU8   = 3
//	IsU16  = 4
//	IsU32  = 5
//	IsU64  = 6
//	IsU128 = 7
//	IsU256 = 8
//	IsI8   = 9
//	IsI16  = 10
//	IsI32  = 11
//	IsI64  = 12
//	IsI128 = 13
//	IsI256 = 14

// Generate and return a go primitive, or a defined type in the case of SCALE primitives that are
// not primitives in go (uint128, uint256, int128, int256)
func (tg *TypeGenerator) GenPrimitive(primitive *types.Si1TypeDefPrimitive, mt *types.PortableTypeV14) (GeneratedType, error) {
	var g GeneratedType
	switch primitive.Si0TypeDefPrimitive {
	case types.IsBool:
		g = &PrimitiveGend{PrimName: "bool", MTy: mt}
	case types.IsChar:
		g = &PrimitiveGend{PrimName: "char", MTy: mt}
	case types.IsStr:
		g = &PrimitiveGend{PrimName: "string", MTy: mt}
	case types.IsU8:
		g = &PrimitiveGend{PrimName: "byte", MTy: mt}
	case types.IsU16:
		g = &PrimitiveGend{PrimName: "uint16", MTy: mt}
	case types.IsU32:
		g = &PrimitiveGend{PrimName: "uint32", MTy: mt}
	case types.IsU64:
		g = &PrimitiveGend{PrimName: "uint64", MTy: mt}
	case types.IsU128:
		// Uint128 is not a primitive in go
		g = &Gend{
			Name: "U128",
			Pkg:  utils.CTYPES,
			MTy:  mt,
		}
	case types.IsU256:
		// Uint256 is not a primitive in go
		g = &Gend{
			Name: "U256",
			Pkg:  utils.CTYPES,
			MTy:  mt,
		}
	case types.IsI8:
		g = &PrimitiveGend{PrimName: "int8", MTy: mt}
	case types.IsI16:
		g = &PrimitiveGend{PrimName: "int16", MTy: mt}
	case types.IsI32:
		g = &PrimitiveGend{PrimName: "int32", MTy: mt}
	case types.IsI64:
		g = &PrimitiveGend{PrimName: "int64", MTy: mt}
	case types.IsI128:
		// int128 is not a primitive in go
		g = &Gend{
			Name: "I128",
			Pkg:  utils.CTYPES,
			MTy:  mt,
		}
	case types.IsI256:
		// int256 is not a primitive in go
		g = &Gend{
			Name: "I256",
			Pkg:  utils.CTYPES,
			MTy:  mt,
		}
	default:
		return nil, fmt.Errorf("unsupported primitive %s", string(primitive.Si0TypeDefPrimitive))
	}
	tg.generated[mt.ID.Int64()] = g
	return g, nil
}
