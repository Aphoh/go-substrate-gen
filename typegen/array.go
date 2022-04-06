package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/tdk"
)

func (tg *TypeGenerator) GenArray(arr *tdk.TDArray, id string) (*gend, error) {
	tyGend, err := tg.GetType(arr.TypeId)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("[%v]%v", arr.Len, tyGend.name)
  g := gend{
		id:   id,
		name: name,
	}
  tg.generated[id] = g
  return &g, nil
}
