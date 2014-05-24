package fixtures

type Something interface {
	DoThings(string, uint64) (int, error)
	DoNothing()
}
