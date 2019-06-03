package fixtures

import (
	"io"
	"net/http"
	. "os"
)

//counterfeiter:generate . DotImports
type DotImports interface {
	DoThings(io.Writer, *File) *http.Client
}
