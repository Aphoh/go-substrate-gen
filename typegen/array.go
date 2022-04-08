package typegen

import (
	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

func (tg *TypeGenerator) GenArray(arr *tdk.TDArray, mt *tdk.MType) (GeneratedType, error) {
	tyGend, err := tg.GetType(arr.TypeId)
	if err != nil {
		return nil, err
	}
	g := ArrayGend{
		Inner: tyGend,
		Len:   arr.Len,
		MTy:   mt,
	}
	tg.generated[mt.Id] = &g
	return &g, nil
}
