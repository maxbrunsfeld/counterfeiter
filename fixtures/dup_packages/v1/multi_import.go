package v1

import (
	"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/v1"
	bv1 "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/v1"
)

type S struct {
}

type MultiAB interface {
	Mine() S
	v1.I
	bv1.I
}
