package fixtures

//go:generate counterfeiter . unexportedFunc
type unexportedFunc func(string, map[string]interface{}) string

//go:generate counterfeiter . unexportedInterface
type unexportedInterface interface {
	Method(string, map[string]interface{}) string
}
