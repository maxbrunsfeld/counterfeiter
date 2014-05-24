package fakes

import "sync"

type FakeSomeInterface struct {
	sync.RWMutex
	DoThingsStub	func(string, uint64) error
	doThingsCalls	[]struct {
		Arg0	string
		Arg1	uint64
	}
	doThingsReturns	struct {
		result0 error
	}
	DoNothingStub	func()
	doNothingCalls	[]struct {
	}
}

func NewFakeSomeInterface() *FakeSomeInterface {
	return &FakeSomeInterface{}
}
func (fake *FakeSomeInterface) DoThings(arg0 string, arg1 uint64) error {
	fake.Lock()
	defer fake.Unlock()
	fake.doThingsCalls = append(fake.doThingsCalls, struct {
		Arg0	string
		Arg1	uint64
	}{arg0, arg1})
	if fake.DoThingsStub != nil {
		return fake.DoThingsStub(arg0, arg1)
	} else {
		return fake.doThingsReturns.result0
	}
}
func (fake *FakeSomeInterface) DoThingsCalls() []struct {
	Arg0	string
	Arg1	uint64
} {
	fake.RLock()
	defer fake.RUnlock()
	return fake.doThingsCalls
}
func (fake *FakeSomeInterface) DoThingsReturns(result0 error) {
	fake.doThingsReturns = struct {
		result0 error
	}{result0: result0}
}
func (fake *FakeSomeInterface) DoNothing() {
	fake.Lock()
	defer fake.Unlock()
	fake.doNothingCalls = append(fake.doNothingCalls, struct {
	}{})
	if fake.DoNothingStub != nil {
		fake.DoNothingStub()
	}
}
func (fake *FakeSomeInterface) DoNothingCalls() []struct {
} {
	fake.RLock()
	defer fake.RUnlock()
	return fake.doNothingCalls
}
