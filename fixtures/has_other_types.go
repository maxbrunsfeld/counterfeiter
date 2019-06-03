package fixtures

//counterfeiter:generate . HasOtherTypes
type HasOtherTypes interface {
	GetThing(SomeString) SomeFunc
}
