package metadata

import (
	"encoding/json"

	"github.com/aphoh/go-substrate-gen/metadata/pal"
	"github.com/aphoh/go-substrate-gen/metadata/tdk"
)

type MetaRoot struct {
	Lookup  MetaLookup      `json:"lookup"`
	Pallets []pal.Pallet      `json:"pallets"`
	Ext     json.RawMessage `json:"extrinsic"`
}

type MetaLookup struct {
	Types []tdk.MType `json:"types"`
}


func ParseMetadata(input []byte) (MetaRoot, error) {
	mr := MetaRoot{}
	err := json.Unmarshal(input, &mr)
	return mr, err
}
