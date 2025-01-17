//go:build !go1.23

package generator

import (
	"go/types"
)

var _ namedType = &typesAlias{}

type typesAlias struct{}

func (a *typesAlias) Obj() *types.TypeName {
	panic("counterfeiter itself must be run/built with go1.23 or newer in order to handle type aliasing")
}

func (a *typesAlias) String() string {
	panic("counterfeiter itself must be run/built with go1.23 or newer in order to handle type aliasing")
}

func (a *typesAlias) TypeArgs() *types.TypeList {
	panic("counterfeiter itself must be run/built with go1.23 or newer in order to handle type aliasing")
}

func (a *typesAlias) Underlying() types.Type {
	panic("counterfeiter itself must be run/built with go1.23 or newer in order to handle type aliasing")
}
