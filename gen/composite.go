package gen

import (
	"fmt"
	"strings"

	"github.com/aphoh/go-substrate-gen/tdk"
	"github.com/dave/jennifer/jen"
)

func (tg *TypeGenerator) GenComposite(v *tdk.TDComposite, id string, path []string) (*gend, error) {
	// name struct id_pathname. Ex: 2_
	sName := path[len(path)-1] + "_" + id
	code := []jen.Code{}
	for i, field := range v.Fields {
		name := field.Name
		if name == "" {
			name = fmt.Sprintf("Field%d", i)
		}
		tyName, err := tg.GetType(field.TypeId)
		_ = tyName
		if err != nil {
			return nil, err
		}
		// Make some comments
		code = append(code, jen.Comment(fmt.Sprintf("Field %d with TypeId=%v", i, field.TypeId)))
		// Add the docs
		for _, d := range field.Docs {
			code = append(code, jen.Comment(d))
		}
		// Add the field
		code = append(code, jen.Id(name).Id(tyName.name))
	}

	// Write new struct with all ids
	tg.f.Comment(fmt.Sprintf("Generated %v with id=%v", strings.Join(path, "_"), id))
	tg.f.Type().Id(sName).Struct(code...)

	g := gend{
		id:   id,
		name: sName,
	}

	tg.generated[id] = g
	return &g, nil
}
