package typegen

import (
	"fmt"
	"strconv"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/dave/jennifer/jen"
)

type GeneratedType interface {
	// Informal name. Is not unique and should only be used to generate other names
	DisplayName() string
	// Actual code that should be rendered when referring to a type
	Code() *jen.Statement
}

// Gend
// This represents a type that lives in a package
type Gend struct {
	Name string
	Pkg  string
}

var _ GeneratedType = &Gend{}

// GlobalName implements GeneratedType
func (eg *Gend) Code() *jen.Statement {
	return jen.Qual(eg.Pkg, eg.Name)
}

// LocalName implements GeneratedType
func (eg *Gend) DisplayName() string {
	return eg.Name
}

// ArrayGend
// Represents an array of the inner type
type ArrayGend struct {
	Len   string
	Inner GeneratedType
}

var _ GeneratedType = &ArrayGend{}

// GlobalName implements GeneratedType
func (ag *ArrayGend) Code() *jen.Statement {
	intLen, err := strconv.Atoi(ag.Len)
	if err != nil {
		panic(fmt.Sprintf("Failed to make array with len=%#v, inner %v", ag.Len, ag.Inner))
	}
	return jen.Index(jen.Lit(intLen)).Custom(utils.TypeOpts, ag.Inner.Code()) // adds an [] to the inner type's value
}

// LocalName implements GeneratedType
func (ag *ArrayGend) DisplayName() string {
	return utils.AsName(ag.Inner.DisplayName(), "Array")
}

// SliceGend
// Represents a slice of the inner type
type SliceGend struct {
	Inner GeneratedType
}

var _ GeneratedType = &SliceGend{}

// GlobalName implements GeneratedType
func (sg *SliceGend) Code() *jen.Statement {
	return jen.Index().Custom(utils.TypeOpts, sg.Inner.Code()) // adds an [] to the inner type's value
}

// LocalName implements GeneratedType
func (sg *SliceGend) DisplayName() string {
	return utils.AsName(sg.Inner.DisplayName(), "Slice")
}

type PrimitiveGend struct {
	PrimName string
}

var _ GeneratedType = &PrimitiveGend{}

// GlobalName implements GeneratedType
func (pg *PrimitiveGend) Code() *jen.Statement {
	return jen.Id(pg.PrimName)
}

// LocalName implements GeneratedType
func (pg *PrimitiveGend) DisplayName() string {
	return utils.AsName(pg.PrimName)
}
