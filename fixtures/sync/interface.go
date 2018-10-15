package sync // import "github.com/maxbrunsfeld/counterfeiter/fixtures/sync"

//go:generate counterfeiter . SyncSomething
type SyncSomething interface {
	DoThings(string, uint64) (int, error)
	DoNothing()
	DoASlice([]byte)
	DoAnArray([4]byte)
}
