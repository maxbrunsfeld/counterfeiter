package genericinterface

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type CustomTypeT any
type CustomTypeU any

// incorrect setup. this would fail
// //counterfeiter:generate . GenericInterfaceBad[T CustomType]
// type GenericInterfaceBad[T CustomTypeT] interface {
// 	ReturnT() T
// 	TakeT(T)
// 	TakeAndReturnT(T) T
// 	DoSomething()
// }

//counterfeiter:generate . GenericInterface
type GenericInterface[T CustomTypeT] interface {
	ReturnT() T
	TakeT(T)
	TakeAndReturnT(T) T
	DoSomething()
}

//counterfeiter:generate . GenericInterfaceAny
type GenericInterfaceAny[T any] interface {
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
