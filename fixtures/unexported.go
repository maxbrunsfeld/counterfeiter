package fixtures

//counterfeiter:generate . unexportedFunc
type unexportedFunc func(string, map[string]interface{}) string

//counterfeiter:generate . unexportedInterface
type unexportedInterface interface {
	Method(string, map[string]interface{}) string
}
