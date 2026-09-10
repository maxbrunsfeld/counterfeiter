package hyphenpackage // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/go-hyphenpackage"

type HyphenType struct{}

// Hyphenated is a method-set constraint declared in a package whose
// directory name (go-hyphenpackage) differs from its package name. A fake
// generated into this directory must use the package name, not the
// directory name.
//
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o . . Hyphenated
type Hyphenated interface {
	Hyphen() string
}
