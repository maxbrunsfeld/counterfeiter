package samefn // import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/samefn"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -o same_fake.go . SomethingFactory

type SomethingFactory func(string, map[string]interface{}) string
