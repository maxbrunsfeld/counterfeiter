package fixtures

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . unexportedFunc
type unexportedFunc func(string, map[string]interface{}) string

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . unexportedInterface
type unexportedInterface interface {
	Method(string, map[string]interface{}) string
}
