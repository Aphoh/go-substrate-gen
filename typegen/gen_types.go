package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

type GeneratedType interface {
	// Informal name. Is not unique and should only be used to generate other names
	DisplayName() string
	// Actual code that should be rendered when referring to a type
	Code() *jen.Statement
	// Parsed type info associated w object
	MType() *types.PortableTypeV14
}

// Gend
// This represents a type that lives in a package
type Gend struct {
	Name string
	Pkg  string
	MTy  *types.PortableTypeV14
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

// Info implements GeneratedType
func (ag *ArrayGend) MType() *types.PortableTypeV14 {
	return ag.MTy
}

// GlobalName implements GeneratedType
func (ag *ArrayGend) Code() *jen.Statement {
	return jen.Index(jen.Lit(ag.Len)).Custom(utils.TypeOpts, ag.Inner.Code()) // adds an [] to the inner type's value
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
	MTy   *types.PortableTypeV14
}

var _ GeneratedType = &SliceGend{}

// Info implements GeneratedType
func (sg *SliceGend) MType() *types.PortableTypeV14 {
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
	MTy      *types.PortableTypeV14
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
func (pg *PrimitiveGend) MType() *types.PortableTypeV14 {
	return pg.MTy
}

type VariantGend struct {
	Gend
	Indices    []uint8
	IsVarNames []string
	AsVarNames [][]string
}

func (vg *VariantGend) IndOf(variantIndex uint8) (int, error) {
	var varI int = -1
	for i := range vg.Indices {
		if vg.Indices[i] == variantIndex {
			varI = i
			break
		}
	}
	if varI == -1 {
		return 0, fmt.Errorf("Unable to find variant %v in gend %v", variantIndex, vg.DisplayName())
	}
	return varI, nil
}
