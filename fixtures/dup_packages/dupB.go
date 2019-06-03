package dup_packages // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages"

import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages/b/foo"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DupB
type DupB interface {
	B() foo.S
}
