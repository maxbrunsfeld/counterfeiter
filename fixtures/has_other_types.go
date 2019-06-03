package fixtures

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . HasOtherTypes
type HasOtherTypes interface {
	GetThing(SomeString) SomeFunc
}
