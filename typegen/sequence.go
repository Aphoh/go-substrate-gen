package typegen

import (
	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

func (tg *TypeGenerator) GenSequence(seq *tdk.TDSequence, mt *tdk.MType) (GeneratedType, error) {
	seqG, err := tg.GetType(seq.TypeId)
	if err != nil {
		return nil, err
	}
	g := &SliceGend{
		Inner: seqG,
		MTy:   mt,
	}

	tg.generated[mt.Id] = g
	return g, nil
}
