package fakes

type FakeSomeInterface struct {
	Method1_ func(arg1 string, arg2 uint64) error
	Method2_ func()
}

func NewFakeSomeInterface() *FakeSomeInterface {
	return &FakeSomeInterface{}
}
func (fake *FakeSomeInterface) Method1(arg1 string, arg2 uint64) error {
	return fake.Method1_(arg1, arg2)
}
func (fake *FakeSomeInterface) Method2() {
	fake.Method2_()
}
