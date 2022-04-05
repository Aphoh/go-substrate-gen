package gen

import (
	"strings"

	"github.com/gobeam/stringy"
)

func asName(strs ...string) string {
	n := stringy.New(strings.Join(strs, "_")).CamelCase(
		"{", "",
		"}", "",
		"[]", "Slice",
    "[", "",
    "]", "")
	return n
}
