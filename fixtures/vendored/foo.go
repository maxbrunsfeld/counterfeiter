package bar

import "apackage"

//go:generate counterfeiter . FooInterface
type FooInterface interface {
	apackage.VendorInterface
}
