package fixtures

type Params struct{}
type Request struct{}
type RequestFactory func(Params, map[string]interface{}) (Request, error)
