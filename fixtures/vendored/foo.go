package bar // import "github.com/maxbrunsfeld/counterfeiter/fixtures/vendored"

import "apackage"

//go:generate counterfeiter . FooInterface
type FooInterface interface {
	apackage.VendorInterface
}
