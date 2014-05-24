package fixtures

type SomeInterface interface {
	DoThings(string, uint64) (int, error)
	DoNothing()
}
