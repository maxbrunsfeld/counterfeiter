package fixtures

//go:generate counterfeiter . SomethingFactory
type SomethingFactory func(string, map[string]interface{}) string
