package fixtures

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . FirstInterface
type FirstInterface interface {
	DoThings()
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SecondInterface
type SecondInterface interface {
	EmbeddedMethod() string
}
