package fixtures

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Something
type Something interface {
	DoThings(string, uint64) (int, error)
	DoNothing()
	DoASlice([]byte)
	DoAnArray([4]byte)
}
