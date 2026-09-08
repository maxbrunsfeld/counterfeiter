package samepackage

// Fakes written into the package that declares the interface must not
// import that package, and may reference unexported names.

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o . . Widget
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o . . gadget

type Thing struct {
	Name string
}

type Widget interface {
	Do(Thing) (Thing, error)
}

// gadget is unexported and has an unexported method, so it can only be
// implemented from inside this package.
type gadget interface {
	Spin(int) error
	stop()
}

// Use exercises both interfaces so the white-box test has something to drive.
func Use(w Widget, g gadget) (string, error) {
	t, err := w.Do(Thing{Name: "in"})
	if err != nil {
		return "", err
	}
	if err := g.Spin(len(t.Name)); err != nil {
		return "", err
	}
	return t.Name, nil
}
