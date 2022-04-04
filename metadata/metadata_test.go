package metadata

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasicMetadata(t *testing.T) {
	raw, err := ioutil.ReadFile("../json-gen/meta.json")
	require.NoError(t, err)

	_, err = ParseMetadata(raw)
	require.NoError(t, err)
}
