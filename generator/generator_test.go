package generator_test

import (
	"github.com/maxbrunsfeld/counterfeiter/locator"

	. "github.com/maxbrunsfeld/counterfeiter/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generator", func() {
	var subject CodeGenerator

	BeforeEach(func() {
		model, _ := locator.GetInterfaceFromFilePath("Something", "../fixtures/something.go")

		subject = CodeGenerator{
			Model:       *model,
			StructName:  "FakeSomething",
			PackageName: "fixturesfakes",
		}
	})

	Describe("generating a fake for a simple interface", func() {
		var fakeFileContents string
		var err error

		BeforeEach(func() {
			fakeFileContents, err = subject.GenerateFake()
		})

		It("should not fail", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should match the correct file contents", func() {
			Expect(fakeFileContents).To(Equal(expectedSimpleFake))
		})
	})

	Describe("generating a fake for a typed function", func() {
		var fakeFileContents string
		var err error

		BeforeEach(func() {
			model, _ := locator.GetInterfaceFromFilePath("RequestFactory", "../fixtures/request_factory.go")

			subject = CodeGenerator{
				Model:       *model,
				StructName:  "FakeRequestFactory",
				PackageName: "fixturesfakes",
			}
			fakeFileContents, err = subject.GenerateFake()
		})

		It("should not fail", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should produce the correct file contents", func() {
			Expect(fakeFileContents).To(Equal(expectedFuncFake))
		})
	})

	Describe("generating a fake for a function return like (a, b int)", func() {
		var fakeFileContents string
		var err error

		BeforeEach(func() {
			model, _ := locator.GetInterfaceFromFilePath("SomethingElse", "../fixtures/compound_return.go")

			subject = CodeGenerator{
				Model:       *model,
				StructName:  "FakeSomethingElse",
				PackageName: "fixturesfakes",
			}
		})

		BeforeEach(func() {
			fakeFileContents, err = subject.GenerateFake()
		})

		It("should not fail", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should match the correct file contents", func() {
			Expect(fakeFileContents).To(Equal(expectedCompoundReturnFake))
		})
	})

})

const expectedSimpleFake string = `// This file was generated by counterfeiter
package fixturesfakes

import (
	"sync"

	"github.com/maxbrunsfeld/counterfeiter/fixtures"
)

type FakeSomething struct {
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
	DoNothingStub        func()
	doNothingMutex       sync.RWMutex
	doNothingArgsForCall []struct{}
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
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSomething) DoThings(arg1 string, arg2 uint64) (int, error) {
	fake.doThingsMutex.Lock()
	fake.doThingsArgsForCall = append(fake.doThingsArgsForCall, struct {
		arg1 string
		arg2 uint64
	}{arg1, arg2})
	fake.recordInvocation("DoThings", []interface{}{arg1, arg2})
	fake.doThingsMutex.Unlock()
	if fake.DoThingsStub != nil {
		return fake.DoThingsStub(arg1, arg2)
	} else {
		return fake.doThingsReturns.result1, fake.doThingsReturns.result2
	}
}

func (fake *FakeSomething) DoThingsCallCount() int {
	fake.doThingsMutex.RLock()
	defer fake.doThingsMutex.RUnlock()
	return len(fake.doThingsArgsForCall)
}

func (fake *FakeSomething) DoThingsArgsForCall(i int) (string, uint64) {
	fake.doThingsMutex.RLock()
	defer fake.doThingsMutex.RUnlock()
	return fake.doThingsArgsForCall[i].arg1, fake.doThingsArgsForCall[i].arg2
}

func (fake *FakeSomething) DoThingsReturns(result1 int, result2 error) {
	fake.DoThingsStub = nil
	fake.doThingsReturns = struct {
		result1 int
		result2 error
	}{result1, result2}
}

func (fake *FakeSomething) DoNothing() {
	fake.doNothingMutex.Lock()
	fake.doNothingArgsForCall = append(fake.doNothingArgsForCall, struct{}{})
	fake.recordInvocation("DoNothing", []interface{}{})
	fake.doNothingMutex.Unlock()
	if fake.DoNothingStub != nil {
		fake.DoNothingStub()
	}
}

func (fake *FakeSomething) DoNothingCallCount() int {
	fake.doNothingMutex.RLock()
	defer fake.doNothingMutex.RUnlock()
	return len(fake.doNothingArgsForCall)
}

func (fake *FakeSomething) DoASlice(arg1 []byte) {
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

func (fake *FakeSomething) DoASliceCallCount() int {
	fake.doASliceMutex.RLock()
	defer fake.doASliceMutex.RUnlock()
	return len(fake.doASliceArgsForCall)
}

func (fake *FakeSomething) DoASliceArgsForCall(i int) []byte {
	fake.doASliceMutex.RLock()
	defer fake.doASliceMutex.RUnlock()
	return fake.doASliceArgsForCall[i].arg1
}

func (fake *FakeSomething) DoAnArray(arg1 [4]byte) {
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

func (fake *FakeSomething) DoAnArrayCallCount() int {
	fake.doAnArrayMutex.RLock()
	defer fake.doAnArrayMutex.RUnlock()
	return len(fake.doAnArrayArgsForCall)
}

func (fake *FakeSomething) DoAnArrayArgsForCall(i int) [4]byte {
	fake.doAnArrayMutex.RLock()
	defer fake.doAnArrayMutex.RUnlock()
	return fake.doAnArrayArgsForCall[i].arg1
}

func (fake *FakeSomething) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.doThingsMutex.RLock()
	defer fake.doThingsMutex.RUnlock()
	fake.doNothingMutex.RLock()
	defer fake.doNothingMutex.RUnlock()
	fake.doASliceMutex.RLock()
	defer fake.doASliceMutex.RUnlock()
	fake.doAnArrayMutex.RLock()
	defer fake.doAnArrayMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeSomething) recordInvocation(key string, args []interface{}) {
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

var _ fixtures.Something = new(FakeSomething)
`

const expectedFuncFake string = `// This file was generated by counterfeiter
package fixturesfakes

import (
	"sync"

	"github.com/maxbrunsfeld/counterfeiter/fixtures"
)

type FakeRequestFactory struct {
	Stub        func(fixtures.Params, map[string]interface{}) (fixtures.Request, error)
	mutex       sync.RWMutex
	argsForCall []struct {
		arg1 fixtures.Params
		arg2 map[string]interface{}
	}
	returns struct {
		result1 fixtures.Request
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRequestFactory) Spy(arg1 fixtures.Params, arg2 map[string]interface{}) (fixtures.Request, error) {
	fake.mutex.Lock()
	fake.argsForCall = append(fake.argsForCall, struct {
		arg1 fixtures.Params
		arg2 map[string]interface{}
	}{arg1, arg2})
	fake.recordInvocation("RequestFactory", []interface{}{arg1, arg2})
	fake.mutex.Unlock()
	if fake.Stub != nil {
		return fake.Stub(arg1, arg2)
	} else {
		return fake.returns.result1, fake.returns.result2
	}
}

func (fake *FakeRequestFactory) CallCount() int {
	fake.mutex.RLock()
	defer fake.mutex.RUnlock()
	return len(fake.argsForCall)
}

func (fake *FakeRequestFactory) ArgsForCall(i int) (fixtures.Params, map[string]interface{}) {
	fake.mutex.RLock()
	defer fake.mutex.RUnlock()
	return fake.argsForCall[i].arg1, fake.argsForCall[i].arg2
}

func (fake *FakeRequestFactory) Returns(result1 fixtures.Request, result2 error) {
	fake.Stub = nil
	fake.returns = struct {
		result1 fixtures.Request
		result2 error
	}{result1, result2}
}

func (fake *FakeRequestFactory) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.mutex.RLock()
	defer fake.mutex.RUnlock()
	return fake.invocations
}

func (fake *FakeRequestFactory) recordInvocation(key string, args []interface{}) {
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

var _ fixtures.RequestFactory = new(FakeRequestFactory).Spy
`
const expectedCompoundReturnFake string = `// This file was generated by counterfeiter
package fixturesfakes

import (
	"sync"

	"github.com/maxbrunsfeld/counterfeiter/fixtures"
)

type FakeSomethingElse struct {
	ReturnStuffStub        func() (a, b int)
	returnStuffMutex       sync.RWMutex
	returnStuffArgsForCall []struct{}
	returnStuffReturns struct {
		result1 int
		result2 int
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeSomethingElse) ReturnStuff() (a, b int) {
	fake.returnStuffMutex.Lock()
	fake.returnStuffArgsForCall = append(fake.returnStuffArgsForCall, struct{}{})
	fake.recordInvocation("ReturnStuff", []interface{}{})
	fake.returnStuffMutex.Unlock()
	if fake.ReturnStuffStub != nil {
		return fake.ReturnStuffStub()
	} else {
		return fake.returnStuffReturns.result1, fake.returnStuffReturns.result2
	}
}

func (fake *FakeSomethingElse) ReturnStuffCallCount() int {
	fake.returnStuffMutex.RLock()
	defer fake.returnStuffMutex.RUnlock()
	return len(fake.returnStuffArgsForCall)
}

func (fake *FakeSomethingElse) ReturnStuffReturns(result1 int, result2 int) {
	fake.ReturnStuffStub = nil
	fake.returnStuffReturns = struct {
		result1 int
		result2 int
	}{result1, result2}
}

func (fake *FakeSomethingElse) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.returnStuffMutex.RLock()
	defer fake.returnStuffMutex.RUnlock()
	return fake.invocations
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
`
