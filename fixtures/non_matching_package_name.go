package fixtures

import (
	different_package_name "github.com/maxbrunsfeld/counterfeiter/fixtures/package_with_different_dirname"
)

type NonMatchingPackageName interface {
	DoThings() different_package_name.SomeStruct
}
