package metadata

import (
	"encoding/json"
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
)

type JRPCResp struct {
	JsonRPC string `json:"jsonrpc"`
	Result  string `json:"result"`
	Id      uint32 `json:"id"`
}

// Returns v14 metadata and the scale-encoded string of the types.Metadata object
func ParseMetadata(input []byte) (*types.MetadataV14, string, error) {
	metaResp := JRPCResp{}
	err := json.Unmarshal(input, &metaResp)
	if err != nil {
		return nil, "", err
	}
	meta := types.Metadata{}
	err = codec.DecodeFromHex(metaResp.Result, &meta)
	if err != nil {
		return nil, "", err
	}
	if meta.Version != 14 {
		return nil, "", fmt.Errorf("Unsupported metadata version: %v, only v14 is currently supported", meta.Version)
	}
	return &meta.AsMetadataV14, metaResp.Result, err
}

type VersResp struct {
	JsonRPC string               `json:"jsonrpc"`
	Result  types.RuntimeVersion `json:"result"`
	Id      uint32               `json:"id"`
}

// Returns the chain version
func ParseVersion(input []byte) (*types.RuntimeVersion, error) {
	versResp := VersResp{}
	err := json.Unmarshal(input, &versResp)
	if err != nil {
		return nil, err
	}
	return &versResp.Result, nil
}
