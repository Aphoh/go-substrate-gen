package gen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/tdk"
)

func (tg *TypeGenerator) GenPrimitive(primitive *tdk.TDPrimitive, id string) (*gend, error) {
	switch *primitive {
	case "U8":
		return &gend{
			name: "byte",
			id:   id,
		}, nil
	default:
		return &gend{}, fmt.Errorf("Unsupported primitive %s", string(*primitive))
	}
}
