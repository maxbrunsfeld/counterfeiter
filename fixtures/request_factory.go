package fixtures

type RequestFactory func(string, map[string]interface{}) (string, error)
