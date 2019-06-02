package fixtures

import (
	"io"
	"net/http"
	. "os"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DotImports
type DotImports interface {
	DoThings(io.Writer, *File) *http.Client
}
