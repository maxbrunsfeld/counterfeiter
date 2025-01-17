//go:build go1.23

package generator

import (
	"go/types"
)

var _ types.Type = &typesAlias{}

type typesAlias = types.Alias
