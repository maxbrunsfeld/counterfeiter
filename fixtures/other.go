package fixtures

//counterfeiter:generate -t . Other
type Other interface {
	DoThings(string, uint64) (int, error)
	DoNothing()
	DoASlice([]byte)
	DoAnArray([4]byte)
}
