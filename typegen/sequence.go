package typegen

import (
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// Generate and return a go slice
func (tg *TypeGenerator) GenSequence(seq *types.Si1TypeDefSequence, mt *types.PortableTypeV14) (GeneratedType, error) {
	seqG, err := tg.GetType(seq.Type.Int64())
	if err != nil {
		return nil, err
	}
	g := &SliceGend{
		Inner: seqG,
		MTy:   mt,
	}

	tg.generated[mt.ID.Int64()] = g
	return g, nil
}
