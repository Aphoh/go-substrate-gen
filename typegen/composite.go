package typegen

import (
	"fmt"
	"strings"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

func (tg *TypeGenerator) GenComposite(v *types.Si1TypeDefComposite, mt *types.PortableTypeV14) (GeneratedType, error) {
	// Handle structs that just wrap, no need to over-wrap
	if len(v.Fields) == 1 {
		f0 := v.Fields[0]
		g, err := tg.GetType(f0.Type.Int64())
		if err != nil {
			return nil, err
		}
		tg.generated[mt.ID.Int64()] = g
		return g, nil
	}

	sName, err := tg.getStructName(mt)
	if err != nil {
		return nil, err
	}
	g := &Gend{
		Name: sName,
		Pkg:  tg.PkgPath,
		MTy:  mt,
	}
	tg.generated[mt.ID.Int64()] = g

	code := []jen.Code{}
	fNameCounts := map[string]uint32{}
	for i, field := range v.Fields {
		code = append(code, jen.Comment(fmt.Sprintf("Field %d with TypeId=%v", i, field.Type.Int64())))
		// Turns out composites don't have unique names... sob
		fc, fname, err := tg.fieldCode(field, "", "", false)
		if err != nil {
			return nil, err
		}
		fNameCounts[fname] += 1
		cnt := fNameCounts[fname]
		if cnt > 1 {
			// Name already exists, generate a new one with a postfix
			fc, fname, err = tg.fieldCode(field, "", fmt.Sprint(cnt-1), false)
			if err != nil {
				return nil, err
			}
		}
		code = append(code, fc...)
	}

	// Write new struct with all ids
	tyPath := utils.PathStrs(mt.Type.Path)
	tg.F.Comment(fmt.Sprintf("Generated %v with id=%v", strings.Join(tyPath, "_"), mt.ID))
	tg.F.Type().Id(sName).Struct(code...)

	return g, nil
}

func (tg *TypeGenerator) fieldCode(f types.Si1Field, prefix, postfix string, useTypeName bool) ([]jen.Code, string, error) {
	var fieldName string
	if f.Name != "" {
		fieldName = string(f.Name)
	} else if useTypeName && f.TypeName != "" {
		fieldName = string(f.TypeName)
	} else {
		fieldName = "Field"
	}
	fieldName = utils.AsName(prefix, fieldName, postfix)

	code := []jen.Code{}

	// Add the docs
	for _, d := range f.Docs {
		code = append(code, jen.Comment(string(d)))
	}

	fieldTy, err := tg.GetType(f.Type.Int64())
	if err != nil {
		return nil, "", err
	}

	// Add the field
	// If it's a rust pointer, use a pointer to avoid recursive structs
	if strings.HasPrefix(string(f.TypeName), "Box") || strings.HasPrefix(string(f.TypeName), "alloc::boxed::Box") || strings.HasPrefix(string(f.TypeName), "OpaqueCall") {
		code = append(code, jen.Id(fieldName).Op("*").Custom(utils.TypeOpts, fieldTy.Code()))
	} else {
		code = append(code, jen.Id(fieldName).Custom(utils.TypeOpts, fieldTy.Code()))
	}

	return code, fieldName, nil
}
