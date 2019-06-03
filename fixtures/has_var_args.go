package fixtures

//counterfeiter:generate . HasVarArgs
type HasVarArgs interface {
	DoThings(int, ...string) int
	DoMoreThings(int, int, ...string) int
}

//counterfeiter:generate . HasVarArgsWithLocalTypes
type HasVarArgsWithLocalTypes interface {
	DoThings(...LocalType)
}

type LocalType struct{}
