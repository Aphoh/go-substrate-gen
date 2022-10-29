package typegen

import (
	"fmt"
	"strings"

	"github.com/aphoh/go-substrate-gen/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/dave/jennifer/jen"
)

// Get the RuntimeCall type for this chain. The RuntimeCall type will always be a variant, and it embeds each pallet's calls.
// example (shortened) output:
// type TemplateRuntimeCall struct {
//   IsSystem                      bool
//   AsSystemField0                FrameSystemPalletCall
//   IsTimestamp                   bool
//   AsTimestampField0             PalletTimestampPalletCall
// }
func (tg *TypeGenerator) GetCallType() (*VariantGend, error) {
	if tg.callId == nil {
		cid, err := getCallTypeId(tg.mtypes)
		if err != nil {
			return nil, err
		}
		tg.callId = &cid
	}

	gend, err := tg.GetType(*tg.callId)
	if err != nil {
		return nil, err
	}
	v, ok := gend.(*VariantGend)
	if !ok {
		return nil, fmt.Errorf("call (id=%v) is not a variant", *tg.callId)
	}
	return v, nil
}

// Generate a function named `AsCall` on the chain runtime calls that converts the chain runtime
// calls into go-substrate-rpc calls by codec-encoding our call.
//
// example output:
//
//	func (c *TemplateRuntimeCall) AsCall() (ret types.Call, err error) {
//		var cb []byte
//		cb, err = codec.Encode(c)
//		if err != nil {
//			return
//		}
//		ret = types.Call{
//			CallIndex: types.CallIndex{
//				SectionIndex: cb[0],
//				MethodIndex:  cb[1],
//			},
//			Args: cb[2:],
//		}
//		return
//}
func (tg *TypeGenerator) GenerateCallHelpers() error {
	callGend, err := tg.GetCallType()
	if err != nil {
		return err
	}

	// The follow three lines output:
	//func (c *RuntimeCall) AsCall() (ctypes.Call, error) {}
	tg.F.Func().Parens(
		jen.Id("c").Op("*").Custom(utils.TypeOpts, callGend.Code()),
	).Id("AsCall").Call().Parens(jen.List(jen.Id("ret").Qual(utils.CTYPES, "Call"), jen.Err().Error())).BlockFunc(func(g1 *jen.Group) {
		g1.Var().Id("cb").Index().Byte() // var cb []byte
		// cb, err = types.EncodeToBytes(c)
		g1.List(jen.Id("cb"), jen.Err()).Op("=").Qual(utils.CCODEC, "Encode").Call(jen.Id("c"))
		utils.ErrorCheckWithNamedArgs(g1)

		var a byte
		var b byte
		_ = types.Call{CallIndex: types.CallIndex{SectionIndex: a, MethodIndex: b}}

		// output:
		// ret = types.Call{
		//			CallIndex: types.CallIndex{
		//				SectionIndex: cb[0],
		//				MethodIndex:  cb[1],
		//			},
		//			Args: cb[2:],
		//		}
		//		return
		g1.Id("ret").Op("=").Qual(utils.CTYPES, "Call").BlockFunc(func(g2 *jen.Group) {
			g2.Id("CallIndex").Op(":").Qual(utils.CTYPES, "CallIndex").BlockFunc(func(g3 *jen.Group) {
				g3.Id("SectionIndex").Op(":").Id("cb").Index(jen.Lit(0)).Op(",")
				g3.Id("MethodIndex").Op(":").Id("cb").Index(jen.Lit(1)).Op(",")
			}).Op(",")
			g2.Id("Args").Op(":").Id("cb").Index(jen.Lit(2).Op(":")).Op(",")
		})
		g1.Return()
	})
	return nil
}

// Get the index of the runtime's call type within the metadata's type array
func getCallTypeId(mtypes map[int64]types.PortableTypeV14) (int64, error) {
	for tyId, ty := range mtypes {
		if len(ty.Type.Path) >= 2 {
			p0 := string(ty.Type.Path[0])
			p1 := string(ty.Type.Path[1])
			// Looking for *_runtime::Call
			if strings.HasSuffix(p0, "_runtime") && p1 == "Call" {
				return tyId, nil
			}
		}
	}
	return 0, fmt.Errorf("no call type found. Expected a path like *_runtime::Call")
}
