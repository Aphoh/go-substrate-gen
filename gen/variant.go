package gen

import (
	"fmt"
	"strings"

	"github.com/aphoh/go-substrate-gen/tdk"
	"github.com/dave/jennifer/jen"
	"github.com/gobeam/stringy"
)

func (tg *TypeGenerator) GenVariant(v *tdk.TDVariant, mt *tdk.MType) (*gend, error) {
	sName := asName(mt.Ty.Path...)

	inner := []jen.Code{}
	for _, variant := range v.Variants {
		inner = append(inner, jen.Id(asName("Is", variant.Name)).Bool())

		for _, f := range variant.Fields {
			fc, err := tg.fieldCode(f, "As"+variant.Name, "")
			if err != nil {
				return nil, err
			}
			inner = append(inner, fc...)
		}

	}

	tg.f.Comment(fmt.Sprintf("Generated %v with id=%v", asName(mt.Ty.Path...), mt.Id))
	tg.f.Type().Id(sName).Struct(inner...)

	g := gend{
		id:   mt.Id,
		name: sName,
	}
	tg.generated[mt.Id] = g
	return &g, nil
}

func asName(strs ...string) string {
	return stringy.New(strings.Join(strs, "_")).CamelCase()
}
