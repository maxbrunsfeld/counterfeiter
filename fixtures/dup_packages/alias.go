package dup_packages // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages"

import (
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages/a"
	afoo "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages/a/foo"
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages/b/foo"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . AliasV1
type AliasV1 interface {
	a.A
	afoo.I
	foo.I
}
