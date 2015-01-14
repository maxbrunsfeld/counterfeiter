package fixtures

type HasVarArgs interface {
	DoThings(int, ...string) int
	DoMoreThings(int, int, ...string) int
}

type HasVarArgsWithLocalTypes interface {
	DoThings(...LocalType)
}

type LocalType struct{}
