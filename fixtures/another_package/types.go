package another_package // import "github.com/maxbrunsfeld/counterfeiter/fixtures/another_package"

type SomeType int

//go:generate counterfeiter . AnotherInterface
type AnotherInterface interface {
	AnotherMethod([]SomeType, map[SomeType]SomeType, *SomeType, SomeType, chan SomeType)
}
