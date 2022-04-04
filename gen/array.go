package gen

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
	return &gend{
		id:   id,
		name: name,
	}, nil
}
