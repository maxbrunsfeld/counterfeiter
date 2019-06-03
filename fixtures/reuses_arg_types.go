package fixtures

//counterfeiter:generate . ReusesArgTypes
type ReusesArgTypes interface {
	DoThings(x, y string)
}
