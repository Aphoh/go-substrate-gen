package gen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/tdk"
	"github.com/dave/jennifer/jen"
)

func (tg *TypeGenerator) GenVariant(v *tdk.TDVariant, mt *tdk.MType) (*gend, error) {
  g, err := tg.getStructName(mt)
  if err != nil {
    return nil, err
  }

	inner := []jen.Code{}
	for _, variant := range v.Variants {
		inner = append(inner, jen.Id(asName("Is", variant.Name)).Bool())

		for i, f := range variant.Fields {
			fc, err := tg.fieldCode(f, "As_"+variant.Name, fmt.Sprint(i))
			if err != nil {
				return nil, err
			}
			inner = append(inner, fc...)
		}

	}

	tg.f.Comment(fmt.Sprintf("Generated %v with id=%v", asName(mt.Ty.Path...), mt.Id))
	tg.f.Type().Id(g.name).Struct(inner...)
	return g, nil
}



