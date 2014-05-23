package fakes

import "sync"

type FakeSomeInterface struct {
	sync.RWMutex
	Method1Stub  func(arg1 string, arg2 uint64) error
	method1Calls []struct {
		Arg1 string
		Arg2 uint64
	}
	Method2Stub  func()
	method2Calls []struct {
	}
}

func NewFakeSomeInterface() *FakeSomeInterface {
	return &FakeSomeInterface{}
}
func (fake *FakeSomeInterface) Method1(arg1 string, arg2 uint64) error {
	fake.Lock()
	defer fake.Unlock()
	fake.method1Calls = append(fake.method1Calls, struct {
		Arg1 string
		Arg2 uint64
	}{arg1, arg2})
	return fake.Method1Stub(arg1, arg2)
}
func (fake *FakeSomeInterface) Method1Calls() []struct {
	Arg1 string
	Arg2 uint64
} {
	fake.RLock()
	defer fake.RUnlock()
	return fake.method1Calls
}
func (fake *FakeSomeInterface) Method2() {
	fake.Lock()
	defer fake.Unlock()
	fake.method2Calls = append(fake.method2Calls, struct {
	}{})
	fake.Method2Stub()
}
func (fake *FakeSomeInterface) Method2Calls() []struct {
} {
	fake.RLock()
	defer fake.RUnlock()
	return fake.method2Calls
}
