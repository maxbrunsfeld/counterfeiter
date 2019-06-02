package some_package // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/hyphenated_package_same_name/some_package"

import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/hyphenated_package_same_name/hyphen-ated/some_package"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SomeInterface
type SomeInterface interface {
	CreateThing() some_package.Thing
}
