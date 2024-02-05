package fixtures

import the_aliased_package "github.com/ikolomiyets/counterfeiter/v6/fixtures/aliased_package"

//counterfeiter:generate . SomethingWithForeignInterface

// SomethingWithForeignInterface is an interface that embeds a foreign interface.
type SomethingWithForeignInterface interface {
	the_aliased_package.InAliasedPackage
}
