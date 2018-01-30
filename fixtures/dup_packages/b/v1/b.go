package v1

type S struct {}

//go:generate counterfeiter . I
type I interface {
	FromB() S
}
