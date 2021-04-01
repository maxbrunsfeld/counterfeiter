package fixtures

//counterfeiter:generate -o ./customfakesdir . CustomOutput
type CustomOutput interface {
	CustomFolder()
}
