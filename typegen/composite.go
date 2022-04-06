package typegen

import (
	"fmt"
	"strings"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/dave/jennifer/jen"
)

func (tg *TypeGenerator) GenComposite(v *tdk.TDComposite, mt *tdk.MType) (*gend, error) {
	// Handle structs that just wrap, no need to over-wrap
	if len(v.Fields) == 1 {
		f0 := v.Fields[0]
		f0gend, err := tg.GetType(f0.TypeId)
		if err != nil {
			return nil, err
		}
		g := gend{
			name: f0gend.name,
			id:   mt.Id,
		}
		tg.generated[mt.Id] = g
		return &g, nil
	}

	g, err := tg.getStructName(mt)
	if err != nil {
		return nil, err
	}

	code := []jen.Code{}
	for i, field := range v.Fields {
		code = append(code, jen.Comment(fmt.Sprintf("Field %d with TypeId=%v", i, field.TypeId)))
		fc, _, err := tg.fieldCode(field, "", fmt.Sprint(i))
		if err != nil {
			return nil, err
		}
		code = append(code, fc...)
	}

	// Write new struct with all ids
	tg.f.Comment(fmt.Sprintf("Generated %v with id=%v", strings.Join(mt.Ty.Path, "_"), mt.Id))
	tg.f.Type().Id(g.name).Struct(code...)

	return g, nil
}

func (tg *TypeGenerator) fieldCode(f tdk.TDField, prefix, postfix string) ([]jen.Code, string, error) {
	fieldName := f.Name
	if fieldName == "" {
		fieldName = "Field"
	}
	fieldName = asName(prefix, fieldName, postfix)

	code := []jen.Code{}

	// Add the docs
	for _, d := range f.Docs {
		code = append(code, jen.Comment(d))
	}

	fieldTy, err := tg.GetType(f.TypeId)
	if err != nil {
		return nil, "", err
	}

	// Add the field
	// If it's a rust pointer, use a pointer to avoid recursive structs
	if strings.HasPrefix(f.TypeName, "Box") {
		code = append(code, jen.Id(fieldName).Op("*").Id(fieldTy.name))
	} else {
		code = append(code, jen.Id(fieldName).Id(fieldTy.name))
	}

	return code, fieldName, nil
}
