package fixtures

//counterfeiter:generate . SomethingFactory
type SomethingFactory func(string, map[string]interface{}) string
