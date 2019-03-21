package dup_packages // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages"

import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages/a/foo"

//go:generate counterfeiter . DupA
type DupA interface {
	A() foo.S
}
