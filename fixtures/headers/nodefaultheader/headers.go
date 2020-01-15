package nodefaultheader // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/headers/nodefaultheader"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . HeaderDefault
type HeaderDefault interface{}

//counterfeiter:generate -header ../specific.header.go.txt . HeaderSpecific
type HeaderSpecific interface{}
