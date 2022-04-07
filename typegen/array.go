package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

func (tg *TypeGenerator) GenArray(arr *tdk.TDArray, id string) (*Gend, error) {
	tyGend, err := tg.GetType(arr.TypeId)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("[%v]%v", arr.Len, tyGend.Name)
  g := Gend{
		Id:   id,
		Name: name,
	}
  tg.generated[id] = g
  return &g, nil
}
