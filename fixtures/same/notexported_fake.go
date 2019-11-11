// Code generated by counterfeiter. DO NOT EDIT.
package same

import (
	"sync"
)

type FakeSomeNotExportedInterface struct {
	DoASliceStub        func([]byte)
	doASliceMutex       sync.RWMutex
	doASliceArgsForCall []struct {
		arg1 []byte
	}
	DoAnArrayStub        func([4]byte)
	doAnArrayMutex       sync.RWMutex
	doAnArrayArgsForCall []struct {
		arg1 [4]byte
	}
	DoNothingStub        func()
	doNothingMutex       sync.RWMutex
	doNothingArgsForCall []struct {
	}
	DoThingsStub        func(string, uint64) (int, error)
	doThingsMutex       sync.RWMutex
	doThingsArgsForCall []struct {
		arg1 string
		arg2 uint64
	}
	doThingsReturns struct {
		result1 int
		result2 error
	}
	doThingsReturnsOnCall map[int]struct {
		result1 int
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSomeNotExportedInterface) DoASlice(arg1 []byte) {
	var arg1Copy []byte
	if arg1 != nil {
		arg1Copy = make([]byte, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.doASliceMutex.Lock()
	fake.doASliceArgsForCall = append(fake.doASliceArgsForCall, struct {
		arg1 []byte
	}{arg1Copy})
	fake.recordInvocation("DoASlice", []interface{}{arg1Copy})
	fake.doASliceMutex.Unlock()
	if fake.DoASliceStub != nil {
		fake.DoASliceStub(arg1)
	}
}

func (fake *FakeSomeNotExportedInterface) DoASliceCallCount() int {
	fake.doASliceMutex.RLock()
	defer fake.doASliceMutex.RUnlock()
	return len(fake.doASliceArgsForCall)
}

func (fake *FakeSomeNotExportedInterface) DoASliceCalls(stub func([]byte)) {
	fake.doASliceMutex.Lock()
	defer fake.doASliceMutex.Unlock()
	fake.DoASliceStub = stub
}

func (fake *FakeSomeNotExportedInterface) DoASliceArgsForCall(i int) []byte {
	fake.doASliceMutex.RLock()
	defer fake.doASliceMutex.RUnlock()
	argsForCall := fake.doASliceArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeSomeNotExportedInterface) DoAnArray(arg1 [4]byte) {
	fake.doAnArrayMutex.Lock()
	fake.doAnArrayArgsForCall = append(fake.doAnArrayArgsForCall, struct {
		arg1 [4]byte
	}{arg1})
	fake.recordInvocation("DoAnArray", []interface{}{arg1})
	fake.doAnArrayMutex.Unlock()
	if fake.DoAnArrayStub != nil {
		fake.DoAnArrayStub(arg1)
	}
}

func (fake *FakeSomeNotExportedInterface) DoAnArrayCallCount() int {
	fake.doAnArrayMutex.RLock()
	defer fake.doAnArrayMutex.RUnlock()
	return len(fake.doAnArrayArgsForCall)
}

func (fake *FakeSomeNotExportedInterface) DoAnArrayCalls(stub func([4]byte)) {
	fake.doAnArrayMutex.Lock()
	defer fake.doAnArrayMutex.Unlock()
	fake.DoAnArrayStub = stub
}

func (fake *FakeSomeNotExportedInterface) DoAnArrayArgsForCall(i int) [4]byte {
	fake.doAnArrayMutex.RLock()
	defer fake.doAnArrayMutex.RUnlock()
	argsForCall := fake.doAnArrayArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeSomeNotExportedInterface) DoNothing() {
	fake.doNothingMutex.Lock()
	fake.doNothingArgsForCall = append(fake.doNothingArgsForCall, struct {
	}{})
	fake.recordInvocation("DoNothing", []interface{}{})
	fake.doNothingMutex.Unlock()
	if fake.DoNothingStub != nil {
		fake.DoNothingStub()
	}
}

func (fake *FakeSomeNotExportedInterface) DoNothingCallCount() int {
	fake.doNothingMutex.RLock()
	defer fake.doNothingMutex.RUnlock()
	return len(fake.doNothingArgsForCall)
}

func (fake *FakeSomeNotExportedInterface) DoNothingCalls(stub func()) {
	fake.doNothingMutex.Lock()
	defer fake.doNothingMutex.Unlock()
	fake.DoNothingStub = stub
}

func (fake *FakeSomeNotExportedInterface) DoThings(arg1 string, arg2 uint64) (int, error) {
	fake.doThingsMutex.Lock()
	ret, specificReturn := fake.doThingsReturnsOnCall[len(fake.doThingsArgsForCall)]
	fake.doThingsArgsForCall = append(fake.doThingsArgsForCall, struct {
		arg1 string
		arg2 uint64
	}{arg1, arg2})
	fake.recordInvocation("DoThings", []interface{}{arg1, arg2})
	fake.doThingsMutex.Unlock()
	if fake.DoThingsStub != nil {
		return fake.DoThingsStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.doThingsReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeSomeNotExportedInterface) DoThingsCallCount() int {
	fake.doThingsMutex.RLock()
	defer fake.doThingsMutex.RUnlock()
	return len(fake.doThingsArgsForCall)
}

func (fake *FakeSomeNotExportedInterface) DoThingsCalls(stub func(string, uint64) (int, error)) {
	fake.doThingsMutex.Lock()
	defer fake.doThingsMutex.Unlock()
	fake.DoThingsStub = stub
}

func (fake *FakeSomeNotExportedInterface) DoThingsArgsForCall(i int) (string, uint64) {
	fake.doThingsMutex.RLock()
	defer fake.doThingsMutex.RUnlock()
	argsForCall := fake.doThingsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeSomeNotExportedInterface) DoThingsReturns(result1 int, result2 error) {
	fake.doThingsMutex.Lock()
	defer fake.doThingsMutex.Unlock()
	fake.DoThingsStub = nil
	fake.doThingsReturns = struct {
		result1 int
		result2 error
	}{result1, result2}
}

func (fake *FakeSomeNotExportedInterface) DoThingsReturnsOnCall(i int, result1 int, result2 error) {
	fake.doThingsMutex.Lock()
	defer fake.doThingsMutex.Unlock()
	fake.DoThingsStub = nil
	if fake.doThingsReturnsOnCall == nil {
		fake.doThingsReturnsOnCall = make(map[int]struct {
			result1 int
			result2 error
		})
	}
	fake.doThingsReturnsOnCall[i] = struct {
		result1 int
		result2 error
	}{result1, result2}
}

func (fake *FakeSomeNotExportedInterface) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.doASliceMutex.RLock()
	defer fake.doASliceMutex.RUnlock()
	fake.doAnArrayMutex.RLock()
	defer fake.doAnArrayMutex.RUnlock()
	fake.doNothingMutex.RLock()
	defer fake.doNothingMutex.RUnlock()
	fake.doThingsMutex.RLock()
	defer fake.doThingsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeSomeNotExportedInterface) recordInvocation(key string, args []interface{}) {
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

var _ someNotExportedInterface = new(FakeSomeNotExportedInterface)
