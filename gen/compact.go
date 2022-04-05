package gen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/tdk"
	"github.com/dave/jennifer/jen"
)

func (tg *TypeGenerator) GenCompact(v *tdk.TDCompact, mt *tdk.MType) (*gend, error) {
	innerT, err := tg.GetType(v.TypeId)
	if err != nil {
		return nil, err
	}
	sName := asName("Compact", innerT.name)

	g := gend{
		name: sName,
		id:   mt.Id,
	}
  tg.generated[mt.Id] = g

	tg.f.Comment(fmt.Sprintf("Generated %v with id=%v", sName, mt.Id))
	tg.f.Type().Id(sName).Struct(
		jen.Comment(fmt.Sprintf("Field %v of id=%v", innerT.name, innerT.id)),
		jen.Id("inner").Id(innerT.name),
	)

  return &g, nil
}
