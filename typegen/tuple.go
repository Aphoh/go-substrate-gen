package typegen

import (
	"fmt"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

// Generate and return a go struct generated from an anonymous rust tuple
func (tg *TypeGenerator) GenTuple(tup *types.Si1TypeDefTuple, mt *types.PortableTypeV14) (GeneratedType, error) {
	var g GeneratedType

	// Empty tuple maps to struct{}
	if len(*tup) == 0 {
		g = &PrimitiveGend{PrimName: "struct{}", MTy: mt}
		tg.generated[mt.ID.Int64()] = g
		return g, nil
	} else if len(*tup) == 1 {
		// Singleton tuples collapse to their contained type
		g, err := tg.GetType((*tup)[0].Int64())
		if err != nil {
			return nil, err
		}
		tg.generated[mt.ID.Int64()] = g
		return g, nil
	}

	var tn string
	if len(*tup) == 2 {
		// Name tuples of only 2 elements based on their interior elements and order
		// TupleOf{Type1Name}{Type2Name}
		g1, err := tg.GetType((*tup)[0].Int64())
		if err != nil {
			return nil, err
		}
		g2, err := tg.GetType((*tup)[1].Int64())
		if err != nil {
			return nil, err
		}
		// Although this name is almost certainly unique, it may still not be totally unique, so we
		// may need to add an integer postfix
		tn = utils.AsName("TupleOf", g1.DisplayName(), g2.DisplayName())

		if tg.nameCount[tn] == 0 {
			tg.nameCount[tn] = 1
		} else {
			tg.nameCount[tn] += 1
			tn = utils.AsName(tn, fmt.Sprint(tg.nameCount[tn]-1))
		}
	} else {
		// We name longer tuples based off their metadata id, as that is guaranteed to be unique, and
		// could help with debugging
		tn = utils.AsName("Tuple", fmt.Sprint(mt.ID.Int64()))
	}

	g = &Gend{
		Name: tn,
		Pkg:  tg.PkgPath,
		MTy:  mt,
	}

	tg.generated[mt.ID.Int64()] = g

	// Generate the tuple definition in the `types/types.go` file

	// Example for mt.ID=121, and the original tuple looks like (int, [32]byte)
	// example output:
	// // Tuple type 121
	// type TupleOfIntByteArray32 struct {
	//   Elem0 int
	//   Elem1 [32]byte
	// }
	code := []jen.Code{}
	for i, te := range *tup {
		ty, err := tg.GetType(te.Int64())
		if err != nil {
			return nil, err
		}
		fName := utils.AsName("Elem", fmt.Sprint(i))
		code = append(code, jen.Id(fName).Custom(utils.TypeOpts, ty.Code()))
	}
	tg.F.Comment(fmt.Sprintf("Tuple type generated from metadata id %v", mt.ID.Int64()))
	tg.F.Type().Id(tn).Struct(code...)
	return g, nil
}
