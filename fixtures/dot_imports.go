package fixtures

import (
	"io"
	"net/http"
	. "os"
)

//go:generate counterfeiter . DotImports
type DotImports interface {
	DoThings(io.Writer, *File) *http.Client
}
