package fixtures

//go:generate counterfeiter . HasOtherTypes
type HasOtherTypes interface {
	GetThing(SomeString) SomeFunc
}
