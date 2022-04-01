package gosubstrategen

type MType struct {
	path   []string
	Params []string               `json:"params"`
	Def    map[string]interface{} `json:"def"`
	Docs   []string               `json:"docs"`
}

// Typedef kinds
const (
	TDKVariant     = "Variant"
	TDKTuple       = "Tuple"
	TDKPrimitive   = "Primitive"
	TDKCompact     = "Compact"
	TDKBitSequence = "BitSequence"
)

type MetaRoot struct {
	Id uint64 `json:"id"`
	Ty MType  `json:"type"`
}

type TDPrimitive string

type TDSequence struct {
	Ty string `json:"type"`
}

