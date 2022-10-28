package utils

import (
	"strings"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
	"github.com/gobeam/stringy"
)

const CTYPES = "github.com/centrifuge/go-substrate-rpc-client/v4/types"
const CCODEC = "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
const GSRPC = "github.com/centrifuge/go-substrate-rpc-client/v4"
const GSRPCState = "github.com/centrifuge/go-substrate-rpc-client/v4/rpc/state"
const TupleIface = "TupleIface"

var TypeOpts = jen.Options{}

// Rules for use in stringy camelcasing. This means that we get rid of all special characters
// listed, except for array indicators, which we replace with a 'Slice'
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

// Replace one or more strings that may contain weird characters with a camelcased, valid go name.
// The first character is capitalized, so it is public
func AsName(strs ...string) string {
	return stringy.New(strings.Join(strs, "_")).CamelCase(rule...)
}

// Replace one or more strings that may contain weird characters with a camelcased, valid go name.
// The first character is lowercased.
func AsArgName(strs ...string) string {
	base := AsName(strs...)
	if len(base) == 0 {
		return ""
	}
	// Lowercase the first char
	base = strings.ToLower(base[:1]) + base[1:]
	return base
}

// Turn a go-substrate-rpc-client path into an array of strings
func PathStrs(p types.Si1Path) (r []string) {
	for _, i := range p {
		r = append(r, string(i))
	}
	return
}

// These functions operate on Jennifer groups. See here for documentation:
// https://github.com/dave/jennifer#func-methods

// output:
// if err != nil {
//   return err
// }
func ErrorCheckG(s *jen.Group) {
	s.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.Err()),
	)
}

// output:
// if err != nil {
//   return nil, err
// }
func ErrorCheckWithNil(s *jen.Group) {
	s.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.List(jen.Nil(), jen.Err())),
	)
}

// output:
// if err != nil {
//   return
// }
func ErrorCheckWithNamedArgs(s *jen.Group) {
	s.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(),
	)
}
