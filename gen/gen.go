package gen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/tdk"
)

func GenType(ty tdk.MType, typeNames map[string]string) (string, error) {
	td, err := ty.Ty.GetTypeDef()
	if err != nil {
		return "", err
	}
	switch v := td.(type) {
	case tdk.TDPrimitive:
		return GenPrimitive(v)
  default:
    return "", fmt.Errorf("Unsupported type %t", v)
	}
}
