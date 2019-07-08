package fixtures

import (
	"context"
	"net/http"
	"time"
)

//counterfeiter:generate . InlineStructParams
type InlineStructParams interface {
	DoSomething(ctx context.Context, body struct {
		SomeString        string
		SomeStringPointer *string
		SomeTime          time.Time
		SomeTimePointer   *time.Time
		HTTPRequest       http.Request
	}) error
}
