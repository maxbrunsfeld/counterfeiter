package defaultheader // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/headers/defaultheader"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -header ../default.header.go.txt -generate

//counterfeiter:generate . HeaderDefault
type HeaderDefault interface{}

//counterfeiter:generate -header ../specific.header.go.txt . HeaderSpecific
type HeaderSpecific interface{}
