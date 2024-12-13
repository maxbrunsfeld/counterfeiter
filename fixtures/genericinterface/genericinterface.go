package genericinterface

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type CustomType any

//counterfeiter:generate . GenericInterface[T CustomType]
type GenericInterface[T CustomType] interface {
	ReturnT() T
	TakeT(T)
	TakeAndReturnT(T) T
	DoSomething()
}

//counterfeiter:generate . GenericInterface2
type GenericInterface2[T CustomType] interface {
	ReturnT() T
	TakeT(T)
	TakeAndReturnT(T) T
	DoSomething()
}
