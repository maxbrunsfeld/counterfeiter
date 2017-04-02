package apackage

type BarType struct{}

type VendorInterface interface {
	BarVendor() BarType
}
