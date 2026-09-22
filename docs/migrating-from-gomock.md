# Migrating from gomock

With [gomock](https://github.com/uber-go/mock) you set up expectations before the code under test runs: which calls, with which arguments, returning what. The mock fails the test on any call that doesn't match. A `counterfeiter` fake only has return values and a record of its calls. Nothing fails until you check the record. Practically, this means that you can use `counterfeiter` fakes without setup and zero values will be returned from any faked method.

The examples use `MySpecialInterface` from the README:

```go
type MySpecialInterface interface {
	DoThings(string, uint64) (int, error)
}
```

## Setup

### `gomock`

```go
ctrl := gomock.NewController(t)
mock := foomocks.NewMockMySpecialInterface(ctrl)
```

### `counterfeiter`
```go
fake := &foofakes.FakeMySpecialInterface{}
```

With `counterfeiter`, there's no controller to initialize. You create a new fake in each test and then use it as needed for that test.

## Assertions

A fake has no assertions of its own. It records its calls, and you check the record with the `testing` package or whichever library your tests already use. The same check, three ways:

### [`testing`](https://pkg.go.dev/testing)

```go
if got := fake.DoThingsCallCount(); got != 1 {
	t.Fatalf("DoThings called %d times, want 1", got)
}
str, num := fake.DoThingsArgsForCall(0)
if str != "stuff" || num != 5 {
	t.Errorf("DoThings called with (%q, %d), want (\"stuff\", 5)", str, num)
}
```

### [`testify`](https://github.com/stretchr/testify)

```go
require.Equal(t, 1, fake.DoThingsCallCount())
str, num := fake.DoThingsArgsForCall(0)
assert.Equal(t, "stuff", str)
assert.Equal(t, uint64(5), num)
```

### [`gomega`](https://onsi.github.io/gomega/)

```go
Expect(fake.DoThingsCallCount()).To(Equal(1))
str, num := fake.DoThingsArgsForCall(0)
Expect(str).To(Equal("stuff"))
Expect(num).To(Equal(uint64(5)))
```

The rest of this page uses gomega, like the README.

## Return values and stubs

| `gomock` | `counterfeiter` |
|---|---|
| `mock.EXPECT().DoThings(gomock.Any(), gomock.Any()).Return(3, nil)` | `fake.DoThingsReturns(3, nil)` |
| two `EXPECT()`s, one per call | `fake.DoThingsReturnsOnCall(0, 3, nil)` then `fake.DoThingsReturnsOnCall(1, 0, err)` |
| `.Do(f)` | `fake.DoThingsCalls(f)` |
| `.DoAndReturn(f)` | `fake.DoThingsCalls(f)` |

`ReturnsOnCall` only covers the calls you give it an entry for. Calls without an entry fall back to `Returns`, or to zero values if you never called that either. When the result depends on the arguments, give the fake a function with `Calls` instead. The function takes precedence over `Returns` and `ReturnsOnCall`, but the `Calls` function will be removed if you call `Returns` or `ReturnsOnCall`. `f` has the method's signature, so a mismatch fails to compile rather than failing at run time the way it does in gomock.

`ReturnsOnCall` picks by call number, not by argument. To answer by argument, the way a gomock test does with one `EXPECT()` per argument value, give the fake a stub with `Calls` and compare the arguments in the stub:

```go
fake.DoThingsCalls(func(s string, _ uint64) (int, error) {
	if s == "a" {
		return 3, nil
	}
	return 0, errBoom
})
```

## Arguments and call counts

A gomock matcher runs when the call happens. With a fake you read the arguments back with `ArgsForCall` once the code under test has returned.

| `gomock` | `counterfeiter` |
|---|---|
| `mock.EXPECT().DoThings("stuff", uint64(5))` | `str, num := fake.DoThingsArgsForCall(0)`, then assert on `str` and `num`; or compare the whole call at once, `Expect(fake.DoThingsArgs()).To(Equal([]foofakes.FakeMySpecialInterfaceDoThingsArgs{{Arg1: "stuff", Arg2: 5}}))` |
| `gomock.Any()` for an argument | leave that argument unchecked |
| `gomock.Not(x)`, `gomock.Len(n)`, `gomock.Regex(s)` | `ArgsForCall` returns a `string` and a `uint64`, so use your assertion library's matcher, for example `Expect(str).To(MatchRegexp("^st"))` |
| `.Times(2)` | `fake.DoThingsCallCount()` is 2 |
| `.AnyTimes()` | not needed. gomock expects exactly one call unless you say otherwise; a fake accepts any number of calls |
| `.MinTimes(n)`, `.MaxTimes(n)` | `Expect(fake.DoThingsCallCount()).To(BeNumerically(">=", n))`, or `"<="` |

## Order of calls

gomock checks order with `InOrder`, across methods and across mocks. Here `other` is a mock of `MyOtherInterface` from the README:

```go
first := mock.EXPECT().DoThings("a", uint64(1)).Return(1, nil)
second := other.EXPECT().DoOtherThings("b", uint64(2)).Return(2, nil)
gomock.InOrder(first, second)
```

A fake records the calls to one method in the order they happened. After two calls to `DoThings`, `ArgsForCall(0)` is the first and `ArgsForCall(1)` the second, and `DoThingsArgs()` returns both in that order:

```go
str, _ := fake.DoThingsArgsForCall(0)
Expect(str).To(Equal("a"))
str, _ = fake.DoThingsArgsForCall(1)
Expect(str).To(Equal("b"))

Expect(fake.DoThingsArgs()).To(Equal([]foofakes.FakeMySpecialInterfaceDoThingsArgs{
	{Arg1: "a", Arg2: 1},
	{Arg1: "b", Arg2: 2},
}))
```

The order of calls to different methods isn't recorded, and neither is the order across two fakes. `Invocations()` is keyed by method name, so it has the same limitation. To check that `DoThings` ran before `DoOtherThings`, give both a stub that appends to the same slice:

```go
var order []string
fake.DoThingsCalls(func(string, uint64) (int, error) {
	order = append(order, "DoThings")
	return 1, nil
})
other.DoOtherThingsCalls(func(string, uint64) (int, error) {
	order = append(order, "DoOtherThings")
	return 2, nil
})

// run the code under test

Expect(order).To(Equal([]string{"DoThings", "DoOtherThings"}))
```

## Unexpected and missing calls

gomock fails a test in two situations. A call that no `EXPECT()` matches fails right away. An `EXPECT()` that never gets its call fails at the end of the test, when the controller's `Finish` runs (`NewController` registers it with `t.Cleanup`). A fake does neither. A call you didn't set up returns zero values, and a call that never happened only shows up if you check `CallCount`. To fail the test if `DoThings` was called:

```go
Expect(fake.DoThingsCallCount()).To(BeZero())
```

To fail it if `DoThings` was never called:

```go
Expect(fake.DoThingsCallCount()).NotTo(BeZero())
```

## Goroutines and types

gomock reports an unmatched call with `t.Fatalf` on the goroutine that made the call. If the code under test called `DoThings` from a goroutine it started, that isn't the test goroutine. `t.Fatalf` off the test goroutine stops the wrong goroutine, and the test can hang, although `go vet` does warns about it. 

A `countefeiter` fake doesn't fail anything when it's called. You check the record from the test goroutine once the goroutine has done its work:

```go
go worker(fake)

Eventually(fake.DoThingsCallCount).Should(Equal(1))
str, _ := fake.DoThingsArgsForCall(0)
Expect(str).To(Equal("stuff"))
```

gomock's `Return` accepts any values, so a wrong return type compiles and is only reported when the test runs:

```go
mock.EXPECT().DoThings(gomock.Any(), gomock.Any()).Return("3", nil)
```

gomock can catch this at compile time if you generate the mock with `mockgen -typed`, which gives `Return`, `Do` and `DoAndReturn` the method's real types. The other recorder methods stay untyped. 

A `counterfeiter` fake is typed throughout: `DoThingsReturns` takes an `int` and an `error`, so this doesn't compile:

```go
fake.DoThingsReturns("3", nil)
```
