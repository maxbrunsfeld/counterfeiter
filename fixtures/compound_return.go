package fixtures

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SomethingElse
type SomethingElse interface {
	ReturnStuff() (a, b int)
}
