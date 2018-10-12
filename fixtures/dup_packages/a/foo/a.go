package foo // import "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/foo"

type S struct{}

//go:generate counterfeiter . I
type I interface {
	FromA() S
}
