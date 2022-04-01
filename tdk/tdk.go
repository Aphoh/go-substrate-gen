package tdk

import (
	"encoding/json"
	"fmt"
)

// Typedef kinds
const (
	TDKArray       = "Array"
	TDKBitSequence = "BitSequence"
	TDKCompact     = "Compact"
	TDKComposite   = "Composite"
	TDKPrimitive   = "Primitive"
	TDKSequence    = "Sequence"
	TDKTuple       = "Tuple"
	TDKVariant     = "Variant"
)

type MType struct {
	Id string    `json:"id"`
	Ty MTypeInfo `json:"type"`
}

type MTypeInfo struct {
	Path   []string                   `json:"path"`
	Params []MTypeParam `json:"params"`
	Def    map[string]json.RawMessage `json:"def"`
	Docs   []string                   `json:"docs"`
}

type MTypeParam struct {
	Name string  `json:"name"`
	Type *string `json:"type"`
}

func (mt *MTypeInfo) GetTypeDef() (interface{}, error) {
	keys := []string{}
	for k := range mt.Def {
		keys = append(keys, k)
	}
	if len(keys) != 1 {
		b, _ := json.Marshal(mt)
		return nil, fmt.Errorf("Multiple or no types in one typedef, %s", string(b))
	}
	typeName := keys[0]
	raw := mt.Def[typeName]

	var res interface{}
	var err error

	switch typeName {
	case TDKArray:
		res = TDArray{}
	case TDKBitSequence:
		panic("Unsupported")
	case TDKCompact:
		res = TDCompact{}
	case TDKComposite:
		res = TDComposite{}
	case TDKPrimitive:
		res = ""
	case TDKSequence:
		res = TDSequence{}
	case TDKTuple:
		res = []TDTuple{}
	case TDKVariant:
		res = TDVariant{}
	default:
		return nil, fmt.Errorf("Unknown type %s", keys[0])
	}

	err = json.Unmarshal(raw, &res)
	return res, err
}

type TDPrimitive string

type TDSequence struct {
	TypeId string `json:"type"`
}

// List of type ids
type TDTuple []string

type TDCompact struct {
	TypeId string `json:"type"`
}
