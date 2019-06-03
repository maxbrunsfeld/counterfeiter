package fixtures

//counterfeiter:generate . FirstInterface
type FirstInterface interface {
	DoThings()
}

//counterfeiter:generate . SecondInterface
type SecondInterface interface {
	EmbeddedMethod() string
}
