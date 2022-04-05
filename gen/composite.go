package gen

import (
	"fmt"
	"strings"

	"github.com/aphoh/go-substrate-gen/tdk"
	"github.com/dave/jennifer/jen"
)

func (tg *TypeGenerator) GenComposite(v *tdk.TDComposite, mt *tdk.MType) (*gend, error) {
	// name struct id_pathname. Ex: 2_
	path := mt.Ty.Path
	id := mt.Id
	sName := asName(mt.Ty.Path...)
	code := []jen.Code{}
	for i, field := range v.Fields {
		code = append(code, jen.Comment(fmt.Sprintf("Field %d with TypeId=%v", i, field.TypeId)))
		fc, err := tg.fieldCode(field, "", fmt.Sprint(i))
		if err != nil {
			return nil, err
		}
		code = append(code, fc...)
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

func (tg *TypeGenerator) fieldCode(f tdk.TDField, prefix, postfix string) ([]jen.Code, error) {
	fieldName := f.Name
	if fieldName == "" {
		fieldName = asName(prefix, "Field", postfix)
	}
	code := []jen.Code{}

	// Add the docs
	for _, d := range f.Docs {
		code = append(code, jen.Comment(d))
	}

	tyName, err := tg.GetType(f.TypeId)
	if err != nil {
		return nil, err
	}

	// Add the field
	code = append(code, jen.Id(fieldName).Id(tyName.name))

	return code, nil
}
