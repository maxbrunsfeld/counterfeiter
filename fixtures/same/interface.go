package same // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/same"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o same_fake.go . SomeInterface
type SomeInterface interface {
	DoThings(string, uint64) (int, error)
	DoNothing()
	DoASlice([]byte)
	DoAnArray([4]byte)
}
