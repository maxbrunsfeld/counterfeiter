package hyphenpackage // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/go-hyphenpackage"

type HyphenType struct{}

// Hyphenated is a method-set constraint declared in a package whose
// directory name (go-hyphenpackage) differs from its package name.
type Hyphenated interface {
	Hyphen() string
}
