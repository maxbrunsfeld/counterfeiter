package internalpkg

import "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/internalpkg/internal"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Context

type Context = internal.Context
