package fixtures

//go:generate counterfeiter . SomethingElse
type SomethingElse interface {
	ReturnStuff() (a, b int)
}
