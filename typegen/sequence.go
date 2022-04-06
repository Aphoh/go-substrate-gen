package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

func (tg *TypeGenerator) GenSequence(seq *tdk.TDSequence, mt *tdk.MType) (*gend, error) {
	seqG, err := tg.GetType(seq.TypeId)
	if err != nil {
		return nil, err
	}
	g := gend{
		id:   mt.Id,
		name: fmt.Sprintf("[]%s", seqG.name),
	}
	tg.generated[mt.Id] = g
	return &g, nil
}
