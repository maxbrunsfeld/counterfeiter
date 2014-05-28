package fixtures

import (
	"io"
	"os"
)

import "net/http"

type HasImports interface {
	DoThings(io.Writer, *os.File) *http.Client
}
