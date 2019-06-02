package fixtures

import (
	"io"
	"net/http"
	some_alias "os"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . HasImports
type HasImports interface {
	DoThings(io.Writer, *some_alias.File) *http.Client
}
