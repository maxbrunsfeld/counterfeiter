package bar

import "apackage"

//go:generate counterfeiter . BarInterface
type BarInterface interface {
	apackage.VendorInterface
}

//go:generate counterfeiter . BarVendoredParameter
type BarVendoredParameter interface {
	Get(*apackage.BarType) *apackage.BarType
}
