package fixtures

type FirstInterface interface {
	DoThings()
}

type SecondInterface interface {
	EmbeddedMethod() string
}

type unexportedInterface interface {
	SomeMethod(string, bool) (int, error)
}
