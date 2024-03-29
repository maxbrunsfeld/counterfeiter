// Code generated by counterfeiter. DO NOT EDIT.
package genericparamfakes

import (
	"sync"

	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/genericparam"
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/genericparam/genericparamtype"
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/genericparam/genericreturntype"
)

type FakeGenericParamFunc struct {
	Stub        func(genericparam.Generic[genericparamtype.T]) genericparam.Generic[genericreturntype.R]
	mutex       sync.RWMutex
	argsForCall []struct {
		arg1 genericparam.Generic[genericparamtype.T]
	}
	returns struct {
		result1 genericparam.Generic[genericreturntype.R]
	}
	returnsOnCall map[int]struct {
		result1 genericparam.Generic[genericreturntype.R]
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGenericParamFunc) Spy(arg1 genericparam.Generic[genericparamtype.T]) genericparam.Generic[genericreturntype.R] {
	fake.mutex.Lock()
	ret, specificReturn := fake.returnsOnCall[len(fake.argsForCall)]
	fake.argsForCall = append(fake.argsForCall, struct {
		arg1 genericparam.Generic[genericparamtype.T]
	}{arg1})
	stub := fake.Stub
	returns := fake.returns
	fake.recordInvocation("GenericParamFunc", []interface{}{arg1})
	fake.mutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return returns.result1
}

func (fake *FakeGenericParamFunc) CallCount() int {
	fake.mutex.RLock()
	defer fake.mutex.RUnlock()
	return len(fake.argsForCall)
}

func (fake *FakeGenericParamFunc) Calls(stub func(genericparam.Generic[genericparamtype.T]) genericparam.Generic[genericreturntype.R]) {
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	fake.Stub = stub
}

func (fake *FakeGenericParamFunc) ArgsForCall(i int) genericparam.Generic[genericparamtype.T] {
	fake.mutex.RLock()
	defer fake.mutex.RUnlock()
	return fake.argsForCall[i].arg1
}

func (fake *FakeGenericParamFunc) Returns(result1 genericparam.Generic[genericreturntype.R]) {
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	fake.Stub = nil
	fake.returns = struct {
		result1 genericparam.Generic[genericreturntype.R]
	}{result1}
}

func (fake *FakeGenericParamFunc) ReturnsOnCall(i int, result1 genericparam.Generic[genericreturntype.R]) {
	fake.mutex.Lock()
	defer fake.mutex.Unlock()
	fake.Stub = nil
	if fake.returnsOnCall == nil {
		fake.returnsOnCall = make(map[int]struct {
			result1 genericparam.Generic[genericreturntype.R]
		})
	}
	fake.returnsOnCall[i] = struct {
		result1 genericparam.Generic[genericreturntype.R]
	}{result1}
}

func (fake *FakeGenericParamFunc) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.mutex.RLock()
	defer fake.mutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeGenericParamFunc) recordInvocation(key string, args []interface{}) {
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

var _ genericparam.GenericParamFunc = new(FakeGenericParamFunc).Spy
