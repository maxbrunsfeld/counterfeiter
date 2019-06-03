package dup_packages // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages"

import (
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages/a/foo"
	bfoo "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages/b/foo"
)

//counterfeiter:generate . AB
type AB interface {
	A() foo.S
	foo.I
	B() bfoo.S
	bfoo.I
}
