package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

func (tg *TypeGenerator) GenSequence(seq *tdk.TDSequence, mt *tdk.MType) (*Gend, error) {
	seqG, err := tg.GetType(seq.TypeId)
	if err != nil {
		return nil, err
	}
	g := Gend{
		Id:   mt.Id,
		Name: fmt.Sprintf("[]%s", seqG.Name),
	}
	tg.generated[mt.Id] = g
	return &g, nil
}
