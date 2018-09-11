package dup_packages

import (
	"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a"
	afoo "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/foo"
	"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo"
)

//go:generate counterfeiter . AliasV1
type AliasV1 interface {
	a.A
	afoo.I
	foo.I
}
