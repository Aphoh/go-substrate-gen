package utils

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/gobeam/stringy"
)

const CTYPES = "github.com/centrifuge/go-substrate-rpc-client/v4/types"
const GSRPC = "github.com/centrifuge/go-substrate-rpc-client/v4"
const TupleEncodeEach = "TupleEncodeEach"

func AsName(strs ...string) string {
	n := stringy.New(strings.Join(strs, "_")).CamelCase(
		"{", "",
		"}", "",
		"[]", "Slice",
		"[", "",
		"]", "")
	return n
}


func ErrorCheckG(s *jen.Group) {
	s.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.Err()),
	)
}
