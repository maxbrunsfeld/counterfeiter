// Code generated by counterfeiter. DO NOT EDIT.
package otherfakes

import (
	"sync"

	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/generate_defaults"
)

type TheSangImposter struct {
	SangStub        func() string
	sangMutex       sync.RWMutex
	sangArgsForCall []struct {
	}
	sangReturns struct {
		result1 string
	}
	sangReturnsOnCall map[int]struct {
		result1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *TheSangImposter) Sang() string {
	fake.sangMutex.Lock()
	ret, specificReturn := fake.sangReturnsOnCall[len(fake.sangArgsForCall)]
	fake.sangArgsForCall = append(fake.sangArgsForCall, struct {
	}{})
	stub := fake.SangStub
	fakeReturns := fake.sangReturns
	fake.recordInvocation("Sang", []interface{}{})
	fake.sangMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *TheSangImposter) SangCallCount() int {
	fake.sangMutex.RLock()
	defer fake.sangMutex.RUnlock()
	return len(fake.sangArgsForCall)
}

func (fake *TheSangImposter) SangCalls(stub func() string) {
	fake.sangMutex.Lock()
	defer fake.sangMutex.Unlock()
	fake.SangStub = stub
}

func (fake *TheSangImposter) SangReturns(result1 string) {
	fake.sangMutex.Lock()
	defer fake.sangMutex.Unlock()
	fake.SangStub = nil
	fake.sangReturns = struct {
		result1 string
	}{result1}
}

func (fake *TheSangImposter) SangReturnsOnCall(i int, result1 string) {
	fake.sangMutex.Lock()
	defer fake.sangMutex.Unlock()
	fake.SangStub = nil
	if fake.sangReturnsOnCall == nil {
		fake.sangReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.sangReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *TheSangImposter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.sangMutex.RLock()
	defer fake.sangMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *TheSangImposter) recordInvocation(key string, args []interface{}) {
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

var _ generate_defaults.Sang = new(TheSangImposter)
