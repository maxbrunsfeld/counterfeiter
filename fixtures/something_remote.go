package fixtures

import "github.com/maxbrunsfeld/counterfeiter/fixtures/aliased_package"

//go:generate counterfeiter . SomethingWithForeignInterface

// SomethingWithForeignInterface is an interface that embeds a foreign interface.
type SomethingWithForeignInterface interface {
	the_aliased_package.InAliasedPackage
}
