package invalid

import (
	"io"
	"net/http"
	some_alias "os"
	nonexistent "this.com/does/not/exist"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . HasValidImports
type HasValidImports interface {
	DoThings(io.Writer, *some_alias.File) *http.Client
}

func ProcessHasValidImports(instance HasValidImports) nonexistent.Report {
	return nonexistent.NewReport().WithHasValidImports(instance)
}
