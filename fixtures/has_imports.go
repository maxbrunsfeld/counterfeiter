package fixtures

import (
	"io"
	"net/http"
	some_alias "os"
)

type HasImports interface {
	DoThings(io.Writer, *some_alias.File) *http.Client
}
