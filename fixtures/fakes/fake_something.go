package fakes

import "sync"

type FakeSomething struct {
	sync.RWMutex
	DoThingsStub	func(string, uint64) (int, error)
	doThingsCalls	[]struct {
		Arg1	string
		Arg2	uint64
	}
	doThingsReturns	struct {
		result1	int
		result2	error
	}
	DoNothingStub	func()
	doNothingCalls	[]struct {
	}
}


func (fake *FakeSomething) DoThings(arg1 string, arg2 uint64) (int, error) {
	fake.Lock()
	defer fake.Unlock()
	fake.doThingsCalls = append(fake.doThingsCalls, struct {
		Arg1	string
		Arg2	uint64
	}{arg1, arg2})
	if fake.DoThingsStub != nil {
		return fake.DoThingsStub(arg1, arg2)
	} else {
		return fake.doThingsReturns.result1, fake.doThingsReturns.result2
	}
}

func (fake *FakeSomething) DoThingsCalls() []struct {
	Arg1	string
	Arg2	uint64
} {
	fake.RLock()
	defer fake.RUnlock()
	return fake.doThingsCalls
}

func (fake *FakeSomething) DoThingsReturns(result1 int, result2 error) {
	fake.doThingsReturns = struct {
		result1	int
		result2	error
	}{result1: result1, result2: result2}
}

func (fake *FakeSomething) DoNothing() {
	fake.Lock()
	defer fake.Unlock()
	fake.doNothingCalls = append(fake.doNothingCalls, struct {
	}{})
	if fake.DoNothingStub != nil {
		fake.DoNothingStub()
	}
}

func (fake *FakeSomething) DoNothingCalls() []struct {
} {
	fake.RLock()
	defer fake.RUnlock()
	return fake.doNothingCalls
}
