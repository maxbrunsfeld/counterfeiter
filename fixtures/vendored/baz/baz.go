package baz

import "apackage"

//go:generate counterfeiter . BazInterface
type BazInterface interface {
	apackage.VendorInterface
}
