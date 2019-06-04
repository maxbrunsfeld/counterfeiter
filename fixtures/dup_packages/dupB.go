package dup_packages // import "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages"

import "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo"

//counterfeiter:generate . DupB
type DupB interface {
	B() foo.S
}
