package fixtures

import the_aliased_package "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/aliased_package"

//go:generate counterfeiter . SomethingWithForeignInterface

// SomethingWithForeignInterface is an interface that embeds a foreign interface.
type SomethingWithForeignInterface interface {
	the_aliased_package.InAliasedPackage
}
