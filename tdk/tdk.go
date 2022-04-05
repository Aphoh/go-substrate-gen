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
	Params []MTypeParam               `json:"params"`
	Def    map[string]json.RawMessage `json:"def"`
	Docs   []string                   `json:"docs"`
}

type MTypeParam struct {
	Name string  `json:"name"`
	Type *string `json:"type"`
}

func (mt *MTypeInfo) GetTypeName() (string, error) {
	keys := []string{}
	for k := range mt.Def {
		keys = append(keys, k)
	}
	if len(keys) != 1 {
		b, _ := json.Marshal(mt)
		return "", fmt.Errorf("Multiple or no types in one typedef, %s", string(b))
	}
	return keys[0], nil
}

func (mt *MTypeInfo) GetArray() (*TDArray, error) {
	var res TDArray
	err := json.Unmarshal(mt.Def[TDKArray], &res)
	return &res, err
}

func (mt *MTypeInfo) GetComposite() (*TDComposite, error) {
	var res TDComposite
	err := json.Unmarshal(mt.Def[TDKComposite], &res)
	return &res, err
}

func (mt *MTypeInfo) GetPrimitive() (*TDPrimitive, error) {
	var res TDPrimitive
	err := json.Unmarshal(mt.Def[TDKPrimitive], &res)
	return &res, err
}

func (mt *MTypeInfo) GetVariant() (*TDVariant, error) {
	var res TDVariant
	err := json.Unmarshal(mt.Def[TDKVariant], &res)
	return &res, err
}

type TDArray struct {
	Len    string `json:"len"`
	TypeId string `json:"type"`
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

type TDField struct {
	Name     string   `json:"name"`
	TypeId   string   `json:"type"` // This is the id of the type that this contains
	TypeName string   `json:"typeName"`
	Index    string   `json:"index"`
	Docs     []string `json:"docs"`
}

type TDComposite struct {
	Fields []TDField `json:"fields"`
}

type TDVariant struct {
	Variants []TDVariantElem `json:"variants"`
}

type TDVariantElem struct {
	Name   string    `json:"name"`
	Fields []TDField `json:"fields"`
}
