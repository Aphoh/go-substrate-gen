package gen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/tdk"
	ctypes "github.com/centrifuge/go-substrate-rpc-client/types"
)

type GenPrimitiveResult struct {
  Name string
}

func GenPrimitive(primitive tdk.TDPrimitive) (GenPrimitiveResult, error) {
	switch primitive {
	case "U8":
    ctypes.NewU8(0)
    return GenPrimitiveResult{
    	Name: "ctypes.U8",
    }, nil
	default:
		return GenPrimitiveResult{}, fmt.Errorf("Unsupported primitive %s", primitive)
	}
}
