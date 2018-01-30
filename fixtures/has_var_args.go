package fixtures

//go:generate counterfeiter . HasVarArgs
type HasVarArgs interface {
	DoThings(int, ...string) int
	DoMoreThings(int, int, ...string) int
}

//go:generate counterfeiter . HasVarArgsWithLocalTypes
type HasVarArgsWithLocalTypes interface {
	DoThings(...LocalType)
}

type LocalType struct{}
