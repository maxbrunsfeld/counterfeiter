Counterfeiter
=============

When writing unit-tests for an object, it is often useful to have fake implementations
of the object's collaborators. In go, such fake implementations cannot be generated
automatically at runtime. This tool allows you to generate them before compiling your code.

### Generating fakes

Choose an interface for which you would like a fake implementation:

```shell
$ cat path/to/some_package/some_interface.go
```

```go
package some_package

type SomeInterface interface {
	DoThings(string, uint64) error
	DoNothing()
}
```

Run counterfeiter like this:

```
$ counterfeiter path/to/some_package SomeInterface
Wrote `FakeSomeInterface` to `path/to/some_package/fakes/fake_some_interface.go`
```

You can customize the location of the ouptut using the `-o` flag, or write the code to standard out by providing `-` as a third argument.

### Using the fake in your tests

Fake objects record their calls:

```go
fake := new(fakes.FakeSomeInterface)

fake.DoThings("stuff", 5)

Expect(fake.DoThingsCalls()).To(HaveLen(1))
Expect(fake.DoThingsCalls()[0].Arg1).To(Equal("stuff"))
Expect(fake.DoThingsCalls()[0].Arg2).To(Equal(uint64(5)))
```

You can set their return values:

```go
fake.DoThingsReturns(errors.New("the-error"))

Expect(fake.DoThings("stuff", 5)).To(Equal(errors.New("the-error")))
```

You can also supply them with stub functions:

```go
fake.DoThingsStub = func(arg1 string, arg2 uint64) error {
	Expect(arg1).To(Equal("stuff"))
	Expect(arg2).To(Equal(uint64(5)))
	return errors.New("the-error")
}

Expect(fake.DoThings("stuff", 5)).To(Equal(errors.New("the-error")))
```

