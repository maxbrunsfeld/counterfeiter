package dup_packages

import (
	"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/v1"
	bv1 "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/v1"
)

type AB interface {
	A() v1.S
	v1.I
	B() bv1.S
	bv1.I
}
