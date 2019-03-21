package dup_packages // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages"

import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages/b/foo"

//go:generate counterfeiter . DupB
type DupB interface {
	B() foo.S
}
