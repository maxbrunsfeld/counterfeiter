package fixtures

//counterfeiter:generate . SomethingElse
type SomethingElse interface {
	ReturnStuff() (a, b int)
}
