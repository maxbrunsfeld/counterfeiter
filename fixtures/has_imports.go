package fixtures

import (
	"net/http"
	"io"
	some_alias "os"
)

type HasImports interface {
	DoThings(io.Writer, *some_alias.File) *http.Client
}
