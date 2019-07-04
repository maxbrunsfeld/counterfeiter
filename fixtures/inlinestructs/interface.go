package inlinestructs

import (
	"context"
	"net/http"
	"time"
)

type SomeInterface interface {
	DoSomething(ctx context.Context, body struct {
		SomeString        string
		SomeStringPointer *string
		SomeTime          time.Time
		SomeTimePointer   *time.Time
		HTTPRequest       http.Request
	}) error
}
