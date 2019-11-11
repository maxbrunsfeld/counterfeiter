package samefn // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/samefn"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o notexported_fake.go . somethingNotExportedFactory

type somethingNotExportedFactory func(string, map[string]interface{}) string
