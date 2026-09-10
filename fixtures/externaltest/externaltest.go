package externaltest

// Fakes generated with -test belong to this package's external test package,
// externaltest_test, and live in _test.go files next to it. They import the
// package that declares the interface, whether that is this package, another
// package in this module, the standard library, or a third-party module.

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -test . Thing
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -test ../samepackage Widget
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -test io.WriteCloser
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -test github.com/onsi/gomega/types.GomegaMatcher

type Thing interface {
	Do(string) (int, error)
}

// Use calls the Thing and is exercised by the black-box test next to it.
func Use(t Thing) (int, error) {
	return t.Do("thing")
}
