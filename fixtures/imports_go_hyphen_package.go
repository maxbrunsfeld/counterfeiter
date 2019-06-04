package fixtures

import (
	"github.com/maxbrunsfeld/counterfeiter/fixtures/go-hyphenpackage"
)

//counterfeiter:generate . ImportsGoHyphenPackage
type ImportsGoHyphenPackage interface {
	UseHyphenType(hyphenpackage.HyphenType)
}
