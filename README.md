Counterfeiter
=============

[![Build Status](https://travis-ci.org/maxbrunsfeld/counterfeiter.svg?branch=master)](https://travis-ci.org/maxbrunsfeld/counterfeiter)

When writing unit-tests for an object, it is often useful to have fake implementations
of the object's collaborators. In go, such fake implementations cannot be generated
automatically at runtime. This tool allows you to generate them before compilation.

### Generating fakes

Choose an interface for which you would like a fake implementation:

```shell
$ cat path/to/foo/something.go
```

```go
package foo

type Something interface {
	DoThings(string, uint64) (int, error)
	DoNothing()
}
```

Run counterfeiter like this:

```shell
$ counterfeiter path/to/foo Something
```

```
Wrote `FakeSomething` to `path/to/foo/foofakes/fake_something.go`
```

You can customize the location of the ouptut using the `-o` flag, or write the code to standard out by providing `-` as a third argument.

If you'd like a fake implementation for an interface you do not own you can do that too by providing only a fully qualified import path:

```shell
$ counterfeiter some/imported/package.Something
```

```
Wrote `FakeSomething` to `current/directory/fakes/fake_something.go`
```

### Using the fake in your tests

Instantiate fakes with `new`:

```go
import "my-repo/path/to/foo/foofakes"

var fake = new(foofakes.FakeSomething)
```

Fakes record the arguments they were called with:

```go
fake.DoThings("stuff", 5)

Expect(fake.DoThingsCallCount()).To(Equal(1))

str, num := fake.DoThingsArgsForCall(0)
Expect(str).To(Equal("stuff"))
Expect(num).To(Equal(uint64(5)))
```

You can set their return values:

```go
fake.DoThingsReturns(3, errors.New("the-error"))

num, err := fake.DoThings("stuff", 5)
Expect(num).To(Equal(3))
Expect(err).To(Equal(errors.New("the-error")))
```

You can set the return value of one or more specific calls:

```go
fake.DoThingsReturnsOnCall(1, 3, errors.New("the-error"))

num, err := fake.DoThings("stuff", 5)
Expect(num).To(Equal(0))
Expect(err).NotTo(HaveOccurred())

num, err = fake.DoThings("stuff", 5)
Expect(num).To(Equal(3))
Expect(err).To(Equal(errors.New("the-error")))
```

You can also supply them with stub functions:

```go
fake.DoThingsStub = func(arg1 string, arg2 uint64) (int, error) {
	Expect(arg1).To(Equal("stuff"))
	Expect(arg2).To(Equal(uint64(5)))
	return 3, errors.New("the-error")
}

num, err := fake.DoThings("stuff", 5)

Expect(num).To(Equal(3))
Expect(err).To(Equal(errors.New("the-error")))
```

### Running counterfeiter's tests

If you want to run counterfeiter's internal tests (perhaps, because you want to contribute a PR), all you have to do is run `scripts/test.sh`.

### Contributions

So you want to contribute to counterfeiter! That's great, here's exactly what you should do:

* open a new github issue, describing your problem, or use case
* help us understand how you want to fix or extend counterfeiter
* write one or more unit tests for the behavior you want
* write the simplest code you can for the feature you're working on
* try to find any opportunities to refactor
* avoid writing code that isn't covered by unit tests

Counterfeiter has a few high level goals for contributors to keep in mind

* keep unit-level test coverage as high as possible
* keep `main.go` as simple as possible
* avoid making the command line options any more complicated
* avoid making the internals of counterfeiter any more complicated

If you have any questions about how to contribute, rest assured that @tjarratt and other maintainers will work with you to ensure we make counterfeiter better, together. This project has largely been maintained by the community, and we greatly appreciate any PR (whether big or small).

### License

Counterfeiter is MIT-licensed.
