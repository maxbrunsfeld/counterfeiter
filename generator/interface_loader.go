package generator

import (
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/types/typeutil"
)

func (f *Fake) addTypesForMethod(sig *types.Signature) {
	for i := 0; i < sig.Results().Len(); i++ {
		ret := sig.Results().At(i)
		f.addImportsFor(ret.Type())
	}
	for i := 0; i < sig.Params().Len(); i++ {
		param := sig.Params().At(i)
		f.addImportsFor(param.Type())
	}
}

func methodForSignature(sig *types.Signature, fakeName string, methodName string, importsMap map[string]Import) Method {
	params := []Param{}
	for i := 0; i < sig.Params().Len(); i++ {
		param := sig.Params().At(i)
		isVariadic := i == sig.Params().Len()-1 && sig.Variadic()
		typ := typeFor(param.Type(), importsMap)
		if isVariadic {
			typ = "..." + typ[2:] // Change []string to ...string
		}
		p := Param{
			Name:       fmt.Sprintf("arg%v", i+1),
			Type:       typ,
			IsVariadic: isVariadic,
			IsSlice:    strings.HasPrefix(typ, "[]"),
		}
		params = append(params, p)
	}
	returns := []Return{}
	for i := 0; i < sig.Results().Len(); i++ {
		ret := sig.Results().At(i)
		r := Return{
			Name: fmt.Sprintf("result%v", i+1),
			Type: typeFor(ret.Type(), importsMap),
		}
		returns = append(returns, r)
	}
	return Method{
		FakeName: fakeName,
		Name:     methodName,
		Returns:  returns,
		Params:   params,
	}
}

func (f *Fake) loadMethodsForInterface() {
	methods := typeutil.IntuitiveMethodSet(f.Target.Type(), nil)
	for i := range methods {
		sig := methods[i].Type().(*types.Signature)
		f.addTypesForMethod(sig)
	}

	importsMap := f.importsMap()
	for i := range methods {
		sig := methods[i].Type().(*types.Signature)
		fun := methods[i].Obj().(*types.Func)
		method := methodForSignature(sig, f.Name, fun.Name(), importsMap)
		f.Methods = append(f.Methods, method)
	}
}
