Counterfeiter
=============

A simple command-line tool that generates your fakes for you.

### Generating fakes

Suppose you have this interface:

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

Then when you run this command:

```
$ counterfeiter path/to/some_package SomeInterface
```

Code for a fake implementation of SomeInterface will be written to `path/to/some_package/fakes/fake_some_interface.go`.
You can configure the location of the output using the `-o` flag.

### Using the fake in your tests

Fakes record their calls:

```go
fake := new(fakes.FakeSomeInterface)

fake.DoThings("stuff", 5)
Expect(fake.DoThingsCalls()).To(HaveLen(1))
Expect(fake.DoThingsCalls()[0].Arg0).To(Equal("stuff"))
Expect(fake.DoThingsCalls()[0].Arg1).To(Equal(uint64(5)))
```

You can set their return values:

```go
fake.DoThingsReturns(errors.New("the-error"))
Expect(fake.DoThings("stuff", 5)).To(Equal(errors.New("the-error")))
```

You can supply the fake with a stub function:

```go
fake.DoThingsStub = func(arg1 string, arg2 uint64) error {
	Expect(arg1).To(Equal("stuff"))
	Expect(arg2).To(Equal(uint64(5)))
	return errors.New("hi")
}

ret := fake.DoThings("stuff", 5)

Expect(ret).To(Equal(errors.New("hi")))
```
