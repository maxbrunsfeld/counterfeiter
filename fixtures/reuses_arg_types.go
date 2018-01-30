package fixtures

//go:generate counterfeiter . ReusesArgTypes
type ReusesArgTypes interface {
	DoThings(x, y string)
}
