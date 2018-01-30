package some_package

import "github.com/maxbrunsfeld/counterfeiter/fixtures/hyphenated_package_same_name/hyphen-ated/some_package"

//go:generate counterfeiter . SomeInterface
type SomeInterface interface {
	CreateThing() some_package.Thing
}
