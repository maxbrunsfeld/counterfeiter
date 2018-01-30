package fixtures

//go:generate counterfeiter . FirstInterface
type FirstInterface interface {
	DoThings()
}

//go:generate counterfeiter . SecondInterface
type SecondInterface interface {
	EmbeddedMethod() string
}
