package fixtures

import (
	"github.com/maxbrunsfeld/counterfeiter/fixtures/aliased_package"
)

type SomethingWithForeignInterface interface {
	the_aliased_package.InAliasedPackage
}
