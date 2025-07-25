// Code generated by counterfeiter. DO NOT EDIT.
package fixturesfakes

import (
	"sync"

	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures"
)

type FakeSomethingElse struct {
	ReturnStuffStub        func() (int, int)
	returnStuffMutex       sync.RWMutex
	returnStuffArgsForCall []struct {
	}
	returnStuffReturns struct {
		result1 int
		result2 int
	}
	returnStuffReturnsOnCall map[int]struct {
		result1 int
		result2 int
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSomethingElse) ReturnStuff() (int, int) {
	fake.returnStuffMutex.Lock()
	ret, specificReturn := fake.returnStuffReturnsOnCall[len(fake.returnStuffArgsForCall)]
	fake.returnStuffArgsForCall = append(fake.returnStuffArgsForCall, struct {
	}{})
	stub := fake.ReturnStuffStub
	fakeReturns := fake.returnStuffReturns
	fake.recordInvocation("ReturnStuff", []interface{}{})
	fake.returnStuffMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeSomethingElse) ReturnStuffCallCount() int {
	fake.returnStuffMutex.RLock()
	defer fake.returnStuffMutex.RUnlock()
	return len(fake.returnStuffArgsForCall)
}

func (fake *FakeSomethingElse) ReturnStuffCalls(stub func() (int, int)) {
	fake.returnStuffMutex.Lock()
	defer fake.returnStuffMutex.Unlock()
	fake.ReturnStuffStub = stub
}

func (fake *FakeSomethingElse) ReturnStuffReturns(result1 int, result2 int) {
	fake.returnStuffMutex.Lock()
	defer fake.returnStuffMutex.Unlock()
	fake.ReturnStuffStub = nil
	fake.returnStuffReturns = struct {
		result1 int
		result2 int
	}{result1, result2}
}

func (fake *FakeSomethingElse) ReturnStuffReturnsOnCall(i int, result1 int, result2 int) {
	fake.returnStuffMutex.Lock()
	defer fake.returnStuffMutex.Unlock()
	fake.ReturnStuffStub = nil
	if fake.returnStuffReturnsOnCall == nil {
		fake.returnStuffReturnsOnCall = make(map[int]struct {
			result1 int
			result2 int
		})
	}
	fake.returnStuffReturnsOnCall[i] = struct {
		result1 int
		result2 int
	}{result1, result2}
}

func (fake *FakeSomethingElse) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeSomethingElse) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ fixtures.SomethingElse = new(FakeSomethingElse)
