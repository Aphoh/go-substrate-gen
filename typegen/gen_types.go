package typegen

import (
	"fmt"
	"strconv"

	"github.com/aphoh/go-substrate-gen/metadata/tdk"
	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/dave/jennifer/jen"
)

type GeneratedType interface {
	// Informal name. Is not unique and should only be used to generate other names
	DisplayName() string
	// Actual code that should be rendered when referring to a type
	Code() *jen.Statement
	// Parsed type info associated w object
	MType() *tdk.MType
	// Is a globally referred to field or does it live in another package
  // E.g. []byte and byte are primitive, ctypes.UCompact is not.
  // This is used to know when to pass things to SCALE by reference v.s. by value
	IsPrimitive() bool
}

// Gend
// This represents a type that lives in a package
type Gend struct {
	Name string
	Pkg  string
	MTy  *tdk.MType
}

// GlobalName implements GeneratedType
func (eg *Gend) Code() *jen.Statement {
	return jen.Qual(eg.Pkg, eg.Name)
}

// LocalName implements GeneratedType
func (eg *Gend) DisplayName() string {
	return eg.Name
}

// TypeInfo implements GeneratedType
func (eg *Gend) MType() *tdk.MType {
	return eg.MTy
}

func (eg *Gend) IsPrimitive() bool {
	return false
}

var _ GeneratedType = &Gend{}

// ArrayGend
// Represents an array of the inner type
type ArrayGend struct {
	Len   string
	Inner GeneratedType
	MTy   *tdk.MType
}

// IsPrimitive implements GeneratedType
func (ag *ArrayGend) IsPrimitive() bool {
	return ag.Inner.IsPrimitive()
}

// Info implements GeneratedType
func (ag *ArrayGend) MType() *tdk.MType {
	return ag.MTy
}

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

var _ GeneratedType = &ArrayGend{}

// SliceGend
// Represents a slice of the inner type
type SliceGend struct {
	Inner GeneratedType
	MTy   *tdk.MType
}

// IsPrimitive implements GeneratedType
func (sg *SliceGend) IsPrimitive() bool {
	return sg.Inner.IsPrimitive()
}

var _ GeneratedType = &SliceGend{}

// Info implements GeneratedType
func (sg *SliceGend) MType() *tdk.MType {
	return sg.MTy
}

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
	MTy      *tdk.MType
}

// IsPrimitive implements GeneratedType
func (*PrimitiveGend) IsPrimitive() bool {
	return true
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

// Info implements GeneratedType
func (pg *PrimitiveGend) MType() *tdk.MType {
	return pg.MTy
}
