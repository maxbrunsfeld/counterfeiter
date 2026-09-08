package fixtures

import (
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/go-hyphenpackage"
)

//counterfeiter:generate . GenericImportedConstraint
type GenericImportedConstraint[T hyphenpackage.Hyphenated] interface {
	ReturnT() T
	DoSomething()
}
