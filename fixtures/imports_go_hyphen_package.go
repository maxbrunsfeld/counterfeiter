package fixtures

import (
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/go-hyphenpackage"
)

//go:generate counterfeiter . ImportsGoHyphenPackage
type ImportsGoHyphenPackage interface {
	UseHyphenType(hyphenpackage.HyphenType)
}
