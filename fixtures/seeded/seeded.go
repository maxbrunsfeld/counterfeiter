package seeded

// The fakes directory already holds a package whose name differs from the
// directory name (see fakes/doc.go). Fakes generated into it join that
// package instead of being named after the directory.

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o fakes . Sower

type Sower interface {
	Sow(seed string) (int, error)
}
