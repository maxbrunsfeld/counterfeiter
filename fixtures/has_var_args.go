package fixtures

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . HasVarArgs
type HasVarArgs interface {
	DoThings(int, ...string) int
	DoMoreThings(int, int, ...string) int
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . HasVarArgsWithLocalTypes
type HasVarArgsWithLocalTypes interface {
	DoThings(...LocalType)
}

type LocalType struct{}
