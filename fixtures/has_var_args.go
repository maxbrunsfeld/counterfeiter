package fixtures

type HasVarArgs interface {
	DoThings(int, ...string) int
}
