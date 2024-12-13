package genericinterface

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type CustomTypeT any
type CustomTypeU any

//counterfeiter:generate . GenericInterface[T CustomType]
type GenericInterface[T CustomTypeT] interface {
	ReturnT() T
	TakeT(T)
	TakeAndReturnT(T) T
	DoSomething()
}

//counterfeiter:generate . GenericInterface2
type GenericInterface2[T CustomTypeT] interface {
	ReturnT() T
	TakeT(T)
	TakeAndReturnT(T) T
	DoSomething()
}

//counterfeiter:generate . GenericInterfaceMultipleTypes
type GenericInterfaceMultipleTypes[T CustomTypeT, U CustomTypeU] interface {
	ReturnT() T
	ReturnU() U
	ReturnTAndU() (T, U)
	TakeT(T)
	TakeU(U)
	TakeTAndU(T, U)
	TakeAndReturnT(T) T
	TakeAndReturnU(U) U
	TakeAndReturnTAndU(T, U) (T, U)
	TakeTAndReturnU(T) U
	DoSomething()
}
