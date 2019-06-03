package fixtures

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SomethingFactory
type SomethingFactory func(string, map[string]interface{}) string
