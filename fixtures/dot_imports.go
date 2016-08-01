package fixtures

import (
	. "bytes"
	"io"
	"net/http"
	. "os"
)

type DotImports interface {
	DoThings(io.Writer, *File) *http.Client
}

func noop() {
	Count(nil, nil)
}
