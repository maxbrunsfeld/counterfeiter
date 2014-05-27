package fakes

import "sync"

type FakeHasVarArgs struct {
	sync.RWMutex
	DoThingsStub		func(int, ...string) int
	doThingsArgsForCall	[]struct {
		arg1	int
		arg2	[]string
	}
	doThingsReturns	struct {
		result1 int
	}
}

func (fake *FakeHasVarArgs) DoThings(arg1 int, arg2 ...string) int {
	fake.Lock()
	defer fake.Unlock()
	fake.doThingsArgsForCall = append(fake.doThingsArgsForCall, struct {
		arg1	int
		arg2	[]string
	}{arg1, arg2})
	if fake.DoThingsStub != nil {
		return fake.DoThingsStub(arg1, arg2...)
	} else {
		return fake.doThingsReturns.result1
	}
}

func (fake *FakeHasVarArgs) DoThingsCallCount() int {
	fake.RLock()
	defer fake.RUnlock()
	return len(fake.doThingsArgsForCall)
}

func (fake *FakeHasVarArgs) DoThingsArgsForCall(i int) (int, []string) {
	fake.RLock()
	defer fake.RUnlock()
	return fake.doThingsArgsForCall[i].arg1, fake.doThingsArgsForCall[i].arg2
}

func (fake *FakeHasVarArgs) DoThingsReturns(result1 int) {
	fake.doThingsReturns = struct {
		result1 int
	}{result1}
}
