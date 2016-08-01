package dup_packages

import (
	"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a"
	av1 "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/v1"
	"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/v1"
)

type AliasV1 interface {
	a.A
	av1.I
	v1.I
}
