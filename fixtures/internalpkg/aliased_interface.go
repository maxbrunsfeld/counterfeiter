package internalpkg

import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/internalpkg/internal"

type Context = internal.Context

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . MyInterface
type MyInterface interface {
	MyFunc(ctx Context)
}
