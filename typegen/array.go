package typegen

import (
	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

func (tg *TypeGenerator) GenArray(arr *tdk.TDArray, id string) (GeneratedType, error) {
	tyGend, err := tg.GetType(arr.TypeId)
	if err != nil {
		return nil, err
	}
	g := ArrayGend{
		Inner: tyGend,
		Len:   arr.Len,
	}
	tg.generated[id] = &g
	return &g, nil
}
