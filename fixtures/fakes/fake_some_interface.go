package fakes

import "sync"

type FakeSomeInterface struct {
	sync.RWMutex
	DoThingsStub	func(string, uint64) error
	doThingsCalls	[]struct {
		Arg1	string
		Arg2	uint64
	}
	doThingsReturns	struct {
		result1 error
	}
	DoNothingStub	func()
	doNothingCalls	[]struct {
	}
}


func NewFakeSomeInterface() *FakeSomeInterface {
	return &FakeSomeInterface{}
}

func (fake *FakeSomeInterface) DoThings(arg1 string, arg2 uint64) error {
	fake.Lock()
	defer fake.Unlock()
	fake.doThingsCalls = append(fake.doThingsCalls, struct {
		Arg1	string
		Arg2	uint64
	}{arg1, arg2})
	if fake.DoThingsStub != nil {
		return fake.DoThingsStub(arg1, arg2)
	} else {
		return fake.doThingsReturns.result1
	}
}

func (fake *FakeSomeInterface) DoThingsCalls() []struct {
	Arg1	string
	Arg2	uint64
} {
	fake.RLock()
	defer fake.RUnlock()
	return fake.doThingsCalls
}

func (fake *FakeSomeInterface) DoThingsReturns(result1 error) {
	fake.doThingsReturns = struct {
		result1 error
	}{result1: result1}
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
