package utils

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/gobeam/stringy"
)

const CTYPES = "github.com/centrifuge/go-substrate-rpc-client/v4/types"
const GSRPC = "github.com/centrifuge/go-substrate-rpc-client/v4"
const GSRPCState = "github.com/centrifuge/go-substrate-rpc-client/v4/rpc/state"
const TupleIface = "TupleIface"

var TypeOpts = jen.Options{}

var rule = []string{"{", "",
	"}", "",
	"[]", "Slice",
	"[", "",
	"]", "",
	">", "",
	"<", "",
	":", "",
	";", "",
	"\n", "",
	",", "",
	" ", "",
	"(", "",
	")", "",
}

func AsName(strs ...string) string {
	return stringy.New(strings.Join(strs, "_")).CamelCase(rule...)
}

func AsArgName(strs ...string) string {
	base := AsName(strs...)
	if len(base) == 0 {
		return ""
	}
	// Lowercase the first char
	base = strings.ToLower(base[:1]) + base[1:]
	return base
}

func ErrorCheckG(s *jen.Group) {
	s.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.Err()),
	)
}

func ErrorCheckWithNil(s *jen.Group) {
	s.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.List(jen.Nil(), jen.Err())),
	)
}

func ErrorCheckWithNamedArgs(s *jen.Group) {
	s.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(),
	)
}
