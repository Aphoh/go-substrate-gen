package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

// A generated type is a wrapper around the code generation for a certain type.
type GeneratedType interface {
	// Informal name. Is not unique and should only be used to generate other names
	DisplayName() string
	// Actual code that should be rendered when referring to a type
	Code() *jen.Statement
	// Get the metadata associated with this type
	MType() *types.PortableTypeV14
}

// Gend
// This represents a type that lives in a package
type Gend struct {
	Name string
	Pkg  string
	MTy  *types.PortableTypeV14
}

// The code name for a generated type is just its fully qualified name
func (eg *Gend) Code() *jen.Statement {
	return jen.Qual(eg.Pkg, eg.Name)
}

// The display name is just the type's small name
func (eg *Gend) DisplayName() string {
	return eg.Name
}

// Get the metadata associated with this type
func (eg *Gend) MType() *types.PortableTypeV14 {
	return eg.MTy
}

var _ GeneratedType = &Gend{}

// ArrayGend
// Represents an array of the inner type
type ArrayGend struct {
	Len   int // use an int so we don't get things like [uint32(0x20)]byte
	Inner GeneratedType
	MTy   *types.PortableTypeV14
}

// Get the metadata associated with this type
func (ag *ArrayGend) MType() *types.PortableTypeV14 {
	return ag.MTy
}

// The code name for an array is an array of the right length of the appropriate type
func (ag *ArrayGend) Code() *jen.Statement {
	return jen.Index(jen.Lit(ag.Len)).Custom(utils.TypeOpts, ag.Inner.Code()) // adds an [len] to the inner type's value
}

// The display name is just InnerTypeArray{Len}
func (ag *ArrayGend) DisplayName() string {
	return utils.AsName(ag.Inner.DisplayName(), "Array", fmt.Sprint(ag.Len))
}

var _ GeneratedType = &ArrayGend{}

// SliceGend
// Represents a slice of the inner type
type SliceGend struct {
	Inner GeneratedType
	MTy   *types.PortableTypeV14
}

var _ GeneratedType = &SliceGend{}

// Get the metadata associated with this type
func (sg *SliceGend) MType() *types.PortableTypeV14 {
	return sg.MTy
}

// The code for a slice is just a slice with the inner type
func (sg *SliceGend) Code() *jen.Statement {
	return jen.Index().Custom(utils.TypeOpts, sg.Inner.Code()) // adds an [] to the inner type's value
}

// The display name is just InnerTypeSlice
func (sg *SliceGend) DisplayName() string {
	return utils.AsName(sg.Inner.DisplayName(), "Slice")
}

// A generated primitive
type PrimitiveGend struct {
	PrimName string
	MTy      *types.PortableTypeV14
}

var _ GeneratedType = &PrimitiveGend{}

// A primitive's code is just itself
func (pg *PrimitiveGend) Code() *jen.Statement {
	return jen.Id(pg.PrimName)
}

// A primitive's local name is itself
func (pg *PrimitiveGend) DisplayName() string {
	return utils.AsName(pg.PrimName)
}

// Get the metadata associated with this type
func (pg *PrimitiveGend) MType() *types.PortableTypeV14 {
	return pg.MTy
}

// A variant is the generated go struct associated with a rust enum / variant.
type VariantGend struct {
	Gend
	// Variant indices are the byte prefix that are used to encode data into the variant. Encoded,
	// the variant is structured like [1 byte index][the data for the corresponding variant]
	Indices []uint8
	// A field for each boolean representing an option of the variant
	IsVarFields []GenField
	//
	AsVarFields [][]GenField
}

// Get the index into the array of variants of a given 1-byte variant prefix
func (vg *VariantGend) IndOf(variantIndex uint8) (int, error) {
	var varI int = -1
	for i := range vg.Indices {
		if vg.Indices[i] == variantIndex {
			varI = i
			break
		}
	}
	if varI == -1 {
		return 0, fmt.Errorf("unable to find variant %v in gend %v", variantIndex, vg.DisplayName())
	}
	return varI, nil
}

type CompositeGend struct {
	Gend
	Fields []GenField
}
