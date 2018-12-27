package generator

import (
	"errors"
	"go/types"
)

func (f *Fake) loadMethodForFunction() error {
	t, ok := f.Target.Type().(*types.Named)
	if !ok {
		return errors.New("target is not a named type")
	}
	sig, ok := t.Underlying().(*types.Signature)
	if !ok {
		return errors.New("target does not have an underlying function signature")
	}
	f.addTypesForMethod(sig)
	importsMap := f.importsMap()
	f.Function = methodForSignature(sig, f.TargetAlias, f.TargetName, importsMap)
	return nil
}
