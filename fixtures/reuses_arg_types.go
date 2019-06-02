package fixtures

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ReusesArgTypes
type ReusesArgTypes interface {
	DoThings(x, y string)
}
