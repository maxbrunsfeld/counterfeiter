package bar

import "apackage"

type BarInterface interface {
	apackage.VendorInterface
}

type BarVendoredParameter interface {
	Get(*apackage.BarType) *apackage.BarType
}
