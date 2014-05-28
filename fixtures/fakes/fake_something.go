package fakes

import "sync"

type FakeSomething struct {
	sync.RWMutex
	DoThingsStub        func(string, uint64) (int, error)
	doThingsArgsForCall []struct {
		arg1 string
		arg2 uint64
	}
	doThingsReturns struct {
		result1 int
		result2 error
	}
	DoNothingStub        func()
	doNothingArgsForCall []struct {
	}
}

func (fake *FakeSomething) DoThings(arg1 string, arg2 uint64) (int, error) {
	fake.Lock()
	defer fake.Unlock()
	fake.doThingsArgsForCall = append(fake.doThingsArgsForCall, struct {
		arg1 string
		arg2 uint64
	}{arg1, arg2})
	if fake.DoThingsStub != nil {
		return fake.DoThingsStub(arg1, arg2)
	} else {
		return fake.doThingsReturns.result1, fake.doThingsReturns.result2
	}
}

func (fake *FakeSomething) DoThingsCallCount() int {
	fake.RLock()
	defer fake.RUnlock()
	return len(fake.doThingsArgsForCall)
}

func (fake *FakeSomething) DoThingsArgsForCall(i int) (string, uint64) {
	fake.RLock()
	defer fake.RUnlock()
	return fake.doThingsArgsForCall[i].arg1, fake.doThingsArgsForCall[i].arg2
}

func (fake *FakeSomething) DoThingsReturns(result1 int, result2 error) {
	fake.doThingsReturns = struct {
		result1 int
		result2 error
	}{result1, result2}
}

func (fake *FakeSomething) DoNothing() {
	fake.Lock()
	defer fake.Unlock()
	fake.doNothingArgsForCall = append(fake.doNothingArgsForCall, struct {
	}{})
	if fake.DoNothingStub != nil {
		fake.DoNothingStub()
	}
}

func (fake *FakeSomething) DoNothingCallCount() int {
	fake.RLock()
	defer fake.RUnlock()
	return len(fake.doNothingArgsForCall)
}
