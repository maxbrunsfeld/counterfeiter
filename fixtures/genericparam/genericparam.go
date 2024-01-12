package genericparam

import (
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/genericparam/genericparamtype"
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/genericparam/genericreturntype"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . GenericParamFunc
type GenericParamFunc func(Generic[genericparamtype.T]) Generic[genericreturntype.R]

//counterfeiter:generate . GenericParamInterface
type GenericParamInterface interface {
	DoSomething(Generic[genericparamtype.T]) Generic[genericreturntype.R]
}

type Generic[T any] struct{ _ T }
