package fixtures

import (
	"io"
	"net/http"
	some_alias "os"
)

//counterfeiter:generate . HasImports
type HasImports interface {
	DoThings(io.Writer, *some_alias.File) *http.Client
}
