package fixtures

import (
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/go-hyphenpackage"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ImportsGoHyphenPackage
type ImportsGoHyphenPackage interface {
	UseHyphenType(hyphenpackage.HyphenType)
}
