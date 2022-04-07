package pal

import "encoding/json"

type Pallet struct {
	Name      string      `json:"name"`
	Storage   Storage     `json:"storage"`
	Calls     STypeLookup `json:"calls"`
	Events    STypeLookup `json:"events"`
	Constants []SConstant `json:"constants"`
	Errors    STypeLookup `json:"errors"`
}

type Storage struct {
	Prefix string  `json:"prefix"`
	Items  []SItem `json:"items"`
}

type SItem struct {
	Name     string                     `json:"name"`
	Modifier string                     `json:"modifier"`
	Type     map[string]json.RawMessage `json:"type"`
	Fallback string                     `json:"fallback"`
	Docs     []string                   `json:"docs"`
}

func (s *SItem) GetTypePlain() (STPlain, error) {
	var v STPlain
	err := json.Unmarshal(s.Type[STKPlain], &v)
	return v, err
}

func (s *SItem) GetTypeMap() (STMap, error) {
	var v STMap
	err := json.Unmarshal(s.Type[STKMap], &v)
	return v, err
}

type STPlain string
type STMap struct {
	Hashers     []string `json:"hashers"`
	KeyTypeId   string   `json:"key"`
	ValueTypeId string   `json:"value"`
}

// Storage type kinds
const (
	STKPlain = "Plain"
	STKMap   = "Map"
)

type STypeLookup struct {
	TypeId string `json:"type"`
}

type SEvents struct {
	TypeId string `json:"type"`
}

type SErrors struct {
	TypeId string `json:"type"`
}

type SConstant struct {
	Name   string   `json:"name"`
	TypeId string   `json:"type"`
	Value  string   `json:"value"`
	Docs   []string `json:"docs"`
}
